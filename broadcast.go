package main

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

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
            log.Printf("Error sending message: %v", err)
            client.Close()
            delete(clients, client)
        }
    }
    clientsLock.Unlock()
}
