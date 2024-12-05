package main

import (
	"log"
	"os"
)

const (
	canvasWidth  = 1024
	canvasHeight = 1024
)

var canvas [canvasHeight][canvasWidth][3]int

func saveCanvas() {
    file, err := os.Create("canvas.bin")
    if err != nil {
        log.Printf("Failed to create canvas file: %v", err)
        return
    }
    defer file.Close()

    buf := make([]byte, canvasWidth*canvasHeight*3)
    for y := 0; y < canvasHeight; y++ {
        for x := 0; x < canvasWidth; x++ {
            buf[(y*canvasWidth+x)*3] = byte(canvas[y][x][0])
            buf[(y*canvasWidth+x)*3+1] = byte(canvas[y][x][1])
            buf[(y*canvasWidth+x)*3+2] = byte(canvas[y][x][2])
        }
    }
    _, err = file.Write(buf)
    if err != nil {
        log.Printf("Failed to write to canvas file: %v", err)
        return
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

    buf := make([]byte, canvasWidth*canvasHeight*3)
    _, err = file.Read(buf)
    if err != nil {
        log.Printf("Failed to read from canvas file: %v", err)
        return
    }

    for y := 0; y < canvasHeight; y++ {
        for x := 0; x < canvasWidth; x++ {
            canvas[y][x] = [3]int{int(buf[(y*canvasWidth+x)*3]), int(buf[(y*canvasWidth+x)*3+1]), int(buf[(y*canvasWidth+x)*3+2])}
        }
    }
    log.Println("Canvas loaded successfully")
}

func placePixel(x, y int, r, g, b int) {
    if x >= 0 && x < canvasWidth && y >= 0 && y < canvasHeight {
        canvas[y][x] = [3]int{r, g, b}
        // Broadcast the pixel data
        dataBroadcast <- Pixel{X: x, Y: y, R: r, G: g, B: b}
    }
}
