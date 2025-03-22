package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"steganografi/internal/api"
)

func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current working directory:", err)
	}

	// Set up static file server
	fs := http.FileServer(http.Dir(filepath.Join(cwd, "web", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up LSB API routes
	http.HandleFunc("/api/lsb/encode/text", api.HandleEncodeText)
	http.HandleFunc("/api/lsb/encode/file", api.HandleEncodeFile)
	http.HandleFunc("/api/lsb/decode/text", api.HandleDecodeText)
	http.HandleFunc("/api/lsb/decode/file", api.HandleDecodeFile)

	// Set up BPCS API routes
	http.HandleFunc("/api/bpcs/encode/text", api.HandleBPCSEncodeText)
	http.HandleFunc("/api/bpcs/encode/file", api.HandleBPCSEncodeFile)
	http.HandleFunc("/api/bpcs/decode/text", api.HandleBPCSDecodeText)
	http.HandleFunc("/api/bpcs/decode/file", api.HandleBPCSDecodeFile)

	// Set up Audio Steganography API routes
	http.HandleFunc("/api/audio/encode/text", api.HandleAudioEncodeText)
	http.HandleFunc("/api/audio/decode/text", api.HandleAudioDecodeText)

	// Set up Video Steganography API routes
	http.HandleFunc("/api/video/encode/text", api.HandleVideoEncodeText)
	http.HandleFunc("/api/video/decode/text", api.HandleVideoDecodeText)
	http.HandleFunc("/api/video/encode/file", api.HandleVideoEncodeFile)
	http.HandleFunc("/api/video/decode/file", api.HandleVideoDecodeFile)

	// Serve the main HTML page
	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(cwd, "web", "index.html"))
	})

	// Serve the audio steganography page
	http.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(cwd, "web", "audio.html"))
	})

	// Add this if you plan to have a video page later
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(cwd, "web", "video.html"))
	})

	// Start the server
	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
