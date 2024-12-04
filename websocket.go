package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Connected wsClients
var wsClients = make(map[*websocket.Conn]bool)
var wsClientsLock sync.Mutex

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
			return true
	},
}

// HandleDataWS handles WebSocket connections
func HandleDataWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
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
		binary.BigEndian.PutUint16(buf[1:3], uint16(end-start))
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
			log.Printf("Error sending canvas data: %v", err)
			conn.Close()
			return
		}
	}

	wsClientsLock.Lock()
	wsClients[conn] = true
	wsClientsLock.Unlock()

	for {
		var pixel Pixel
		err := conn.ReadJSON(&pixel)
		if err != nil {
			wsClientsLock.Lock()
			delete(wsClients, conn)
			wsClientsLock.Unlock()
			break
		}
		dataBroadcast <- pixel
	}
}
