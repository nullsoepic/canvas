package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

// FRONTEND ENDPOINTS

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

// DRAWING ENDPOINTS

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
	
	placePixel(data.X, data.Y, data.R, data.G, data.B)
	w.WriteHeader(http.StatusOK)
}


func GetPixel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")
	
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