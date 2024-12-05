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
)

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
		go func() {
			drawingMutex.Lock()
			defer drawingMutex.Unlock()
			placePixel(data.X, data.Y, data.R, data.G, data.B)
			conn.WriteMessage(websocket.TextMessage, []byte("ok"))
		}()
	}
}
