package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var canvas [512][512][3]int // 2D array to store RGB values for each pixel

// Pixel structure for JSON response
type Pixel struct {
    X int `json:"x"`
    Y int `json:"y"`
    R int `json:"r"`
    G int `json:"g"`
    B int `json:"b"`
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// Connected clients
var clients = make(map[*websocket.Conn]bool)
var clientsLock sync.Mutex

// Broadcast channel
var broadcast = make(chan Pixel)

// UpdatePixel updates the color of a specific pixel and broadcasts the update
func UpdatePixel(w http.ResponseWriter, r *http.Request) {
    var data Pixel
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if data.X < 0 || data.X >= 512 || data.Y < 0 || data.Y >= 512 {
        http.Error(w, "Invalid coordinates", http.StatusBadRequest)
        return
    }
    canvas[data.Y][data.X] = [3]int{data.R, data.G, data.B}
    broadcast <- data
    w.WriteHeader(http.StatusOK)
}

// GetPixel retrieves the color of a specific pixel
func GetPixel(w http.ResponseWriter, r *http.Request) {
    xStr := r.URL.Query().Get("x")
    yStr := r.URL.Query().Get("y")
    
    // Convert x and y to int
    x, errX := strconv.Atoi(xStr)
    y, errY := strconv.Atoi(yStr)
    
    if errX != nil || errY != nil || x < 0 || x >= 512 || y < 0 || y >= 512 {
        http.Error(w, "Invalid coordinates", http.StatusBadRequest)
        return
    }
    pixel := canvas[y][x]
    json.NewEncoder(w).Encode(Pixel{R: pixel[0], G: pixel[1], B: pixel[2]})
}

// ServeHTML serves the HTML page with the canvas
func ServeHTML(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("static/index.html")
    if err != nil {
        http.Error(w, "Error loading template", http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}

// ServeDocs serves the HTML page with the documentation for the api
func ServeDocs(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("static/docs.html")
    if err != nil {
        http.Error(w, "Error loading template", http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
			log.Println(err)
			return
	}
	defer conn.Close()

	pixelsPerMessage := 5000
	totalPixels := 512 * 512
	numMessages := (totalPixels + pixelsPerMessage - 1) / pixelsPerMessage

	for i := 0; i < numMessages; i++ {
			start := i * pixelsPerMessage
			end := start + pixelsPerMessage
			if end > totalPixels {
					end = totalPixels
			}
			buf := make([]byte, 1+2+(pixelsPerMessage*7))
			buf[0] = 0x01
			binary.BigEndian.PutUint16(buf[1:3], uint16(end - start))
			offset := 3
			for y := 0; y < 512; y++ {
					for x := 0; x < 512; x++ {
							idx := y*512 + x
							if idx < start || idx >= end {
									continue
							}
							binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(x))
							binary.BigEndian.PutUint16(buf[offset+2:offset+4], uint16(y))
							buf[offset+4] = uint8(canvas[y][x][0])
							buf[offset+5] = uint8(canvas[y][x][1])
							buf[offset+6] = uint8(canvas[y][x][2])
							offset += 7
					}
			}
			err := conn.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
					log.Printf("error sending canvas data: %v", err)
					conn.Close()
					return
			}
	}

	clientsLock.Lock()
	clients[conn] = true
	clientsLock.Unlock()

	for {
			var pixel Pixel
			err := conn.ReadJSON(&pixel)
			if err != nil {
					clientsLock.Lock()
					delete(clients, conn)
					clientsLock.Unlock()
					break
			}
			broadcast <- pixel
	}
}

var broadcastBatchSize = 100
var broadcastTimeout = 30 * time.Millisecond

func handleMessages() {
    var pixels []Pixel
    ticker := time.NewTicker(broadcastTimeout)
    defer ticker.Stop()
    for {
        select {
        case pixel := <-broadcast:
            pixels = append(pixels, pixel)
            if len(pixels) >= broadcastBatchSize {
                sendBatch(pixels, 0x00)
                pixels = nil
            }
        case <-ticker.C:
            if len(pixels) > 0 {
                sendBatch(pixels, 0x00)
                pixels = nil
            }
        }
    }
}

func sendBatch(pixels []Pixel, msgType byte) {
    buf := make([]byte, 1+2+len(pixels)*7)
    buf[0] = msgType
    binary.BigEndian.PutUint16(buf[1:3], uint16(len(pixels)))
    offset := 3
    for _, pixel := range pixels {
        binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(pixel.X))
        binary.BigEndian.PutUint16(buf[offset+2:offset+4], uint16(pixel.Y))
        buf[offset+4] = uint8(pixel.R)
        buf[offset+5] = uint8(pixel.G)
        buf[offset+6] = uint8(pixel.B)
        offset += 7
    }
    clientsLock.Lock()
    for client := range clients {
        err := client.WriteMessage(websocket.BinaryMessage, buf)
        if err != nil {
            log.Printf("error sending message: %v", err)
            client.Close()
            delete(clients, client)
        }
    }
    clientsLock.Unlock()
}

func saveCanvas() {
    file, err := os.Create("canvas.bin")
    if err != nil {
        log.Printf("Failed to create canvas file: %v", err)
        return
    }
    defer file.Close()

    for y := 0; y < 512; y++ {
        for x := 0; x < 512; x++ {
            _, err := file.Write([]byte{byte(canvas[y][x][0]), byte(canvas[y][x][1]), byte(canvas[y][x][2])})
            if err != nil {
                log.Printf("Failed to write to canvas file: %v", err)
                return
            }
        }
    }
    log.Println("Canvas saved successfully")
}

func loadCanvas() {
    file, err := os.Open("canvas.bin")
    if err != nil {
        log.Printf("Failed to open canvas file: %v", err)
        return
    }
    defer file.Close()

    buf := make([]byte, 512*512*3)
    _, err = file.Read(buf)
    if err != nil {
        log.Printf("Failed to read from canvas file: %v", err)
        return
    }

    for y := 0; y < 512; y++ {
        for x := 0; x < 512; x++ {
            canvas[y][x] = [3]int{int(buf[(y*512+x)*3]), int(buf[(y*512+x)*3+1]), int(buf[(y*512+x)*3+2])}
        }
    }
    log.Println("Canvas loaded successfully")
}

func main() {
    loadCanvas()

    http.HandleFunc("/", ServeHTML)
    http.HandleFunc("/docs", ServeDocs)
    http.HandleFunc("/updatePixel", UpdatePixel)
    http.HandleFunc("/getPixel", GetPixel)
    http.HandleFunc("/ws", HandleWebSocket)

    go handleMessages()

    // Save canvas every minute
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        for range ticker.C {
            saveCanvas()
        }
    }()

    var port int = 9992

    server := &http.Server{
        Addr: fmt.Sprintf(":%d", port),
        Handler: http.DefaultServeMux,
    }

    // Handle OS signals for graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-stop
        log.Println("Shutting down server...")
        saveCanvas()
        if err := server.Shutdown(context.Background()); err != nil {
            log.Printf("Server shutdown failed: %v", err)
        }
        log.Println("Server stopped")
        os.Exit(0)
    }()

    fmt.Printf("Server is running on http://127.0.0.1:%d/\n", port)
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Printf("Server failed: %v", err)
    }
}