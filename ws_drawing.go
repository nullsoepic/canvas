package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	wsDrawingUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	drawingMutex sync.Mutex
	pixelQueue   = make(chan Pixel, 1000)
)

func pixelWorker() {
	for pixel := range pixelQueue {
		drawingMutex.Lock()
		placePixel(pixel.X, pixel.Y, pixel.R, pixel.G, pixel.B)
		drawingMutex.Unlock()
	}
}

func init() {
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go pixelWorker()
	}
}

func HandleDrawWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsDrawingUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		log.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer conn.Close()
	log.Println("WebSocket connection upgraded successfully")

	for {
		var data Pixel
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Println("Error reading message:", err)
			conn.WriteMessage(websocket.TextMessage, []byte("err"))
			break
		}
		if data.X < 0 || data.X >= canvasWidth || data.Y < 0 || data.Y >= canvasHeight {
			log.Println("Invalid pixel data:", data)
			conn.WriteMessage(websocket.TextMessage, []byte("err"))
			continue
		}
		
		pixelQueue <- data
		conn.WriteMessage(websocket.TextMessage, []byte("ok"))
	}
}
