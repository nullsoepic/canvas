package main

import (
	"log"
	"os"
)

var canvas [512][512][3]int // 2D array to store RGB values for each pixel

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

func placePixel(x, y int, r, g, b int) {
    if x >= 0 && x < 512 && y >= 0 && y < 512 {
        canvas[y][x] = [3]int{r, g, b}
    } else {
        log.Printf("Pixel coordinates out of bounds: (%d, %d)", x, y)
    }
}
