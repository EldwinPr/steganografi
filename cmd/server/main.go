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

	// Set up API routes
	http.HandleFunc("/api/encode", api.HandleEncode)
	http.HandleFunc("/api/decode", api.HandleDecode)

	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(cwd, "web", "templates", "index.html"))
	})

	// Start the server
	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
