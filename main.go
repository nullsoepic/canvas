package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Pixel structure for JSON response
type Pixel struct {
    X int `json:"x"`
    Y int `json:"y"`
    R int `json:"r"`
    G int `json:"g"`
    B int `json:"b"`
}

func main() {
    loadCanvas()
    
    http.HandleFunc("/", ServeIndex)
    http.HandleFunc("/docs", ServeDocs)
    http.HandleFunc("/updatePixel", UpdatePixel)
    http.HandleFunc("/getPixel", GetPixel)
    http.HandleFunc("/ws", HandleDataWS)
    http.HandleFunc("/cws", WebSocketUpdatePixel)

    go handleMessages()

    // Save canvas every minute
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        for range ticker.C {
            saveCanvas()
        }
    }()

    var port int = 9999

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
