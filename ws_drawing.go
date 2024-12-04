package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var wsDrawingUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleDrawWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsDrawingUpgrader.Upgrade(w, r, nil)
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
		conn.WriteMessage(websocket.TextMessage, []byte("ok"))
		inactivityTimer.Reset(30 * time.Second)
	}

	inactivityTimer.Stop()
}