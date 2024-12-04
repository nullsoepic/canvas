package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpdatePixelUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeIndex serves the HTML page with the canvas
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// ServeDocs serves the HTML page with the documentation for the api
func ServeDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("static/docs.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// UpdatePixel updates the color of a specific pixel and broadcasts the update
func UpdatePixel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data Pixel
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	if data.X < 0 || data.X >= 512 || data.Y < 0 || data.Y >= 512 {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}
	// Use placePixel to update the canvas
	placePixel(data.X, data.Y, data.R, data.G, data.B)
	dataBroadcast <- data
	w.WriteHeader(http.StatusOK)
}

// GetPixel retrieves the color of a specific pixel
func GetPixel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")
	
	// Convert x and y to int
	x, errX := strconv.Atoi(xStr)
	y, errY := strconv.Atoi(yStr)
	
	if errX != nil || errY != nil {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	if x < 0 || x >= 512 || y < 0 || y >= 512 {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}
	pixel := canvas[y][x]
	json.NewEncoder(w).Encode(Pixel{R: pixel[0], G: pixel[1], B: pixel[2]})
}

// WebSocketUpdatePixel handles websocket connections for updating pixels
func WebSocketUpdatePixel(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpdatePixelUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Set a timer to close the connection after 30 seconds of inactivity
	inactivityTimer := time.AfterFunc(30*time.Second, func() {
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
	})

	for {
		var data Pixel
		err := conn.ReadJSON(&data)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("err"))
			break
		}
		if data.X < 0 || data.X >= 512 || data.Y < 0 || data.Y >= 512 {
			conn.WriteMessage(websocket.TextMessage, []byte("err"))
			continue
		}
		placePixel(data.X, data.Y, data.R, data.G, data.B)
		dataBroadcast <- data
		conn.WriteMessage(websocket.TextMessage, []byte("ok"))
		inactivityTimer.Reset(30 * time.Second)
	}

	inactivityTimer.Stop()
}
