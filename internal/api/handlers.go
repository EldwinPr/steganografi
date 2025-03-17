package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"steganografi/internal/steganography"
)

// Response represents the API response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// HandleEncode handles the encoding of a message into an image
func HandleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")
	message := r.FormValue("message")
	bitsUsedStr := r.FormValue("bitsUsed")

	bitsUsed, err := strconv.Atoi(bitsUsedStr)
	if err != nil || bitsUsed < 1 || bitsUsed > 3 {
		bitsUsed = 1 // Default to 1 bit if invalid
	}

	// Get the file from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		sendErrorResponse(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input and output file paths
	inputPath := filepath.Join(tempDir, "input_"+timestamp+filepath.Ext(handler.Filename))
	outputPath := filepath.Join(tempDir, "output_"+timestamp+".png")

	// Save the uploaded file
	inputFile, err := os.Create(inputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	defer inputFile.Close()
	defer os.Remove(inputPath) // Clean up

	_, err = io.Copy(inputFile, file)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	inputFile.Close() // Close now so it can be read

	// Create LSB encoder
	encoder, err := steganography.NewLSBEncoder(seed, bitsUsed)
	if err != nil {
		sendErrorResponse(w, "Failed to create encoder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the message
	err = encoder.EncodeMessage(inputPath, outputPath, message)
	if err != nil {
		sendErrorResponse(w, "Failed to encode message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath) // Clean up

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=stego_image.png")
	w.Header().Set("Content-Type", "image/png")

	// Send the file
	outputFile, err := os.Open(outputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to read output file", http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(w, outputFile)
	if err != nil {
		sendErrorResponse(w, "Failed to send output file", http.StatusInternalServerError)
		return
	}
}

// HandleDecode handles the decoding of a message from an image
func HandleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")
	bitsUsedStr := r.FormValue("bitsUsed")

	bitsUsed, err := strconv.Atoi(bitsUsedStr)
	if err != nil || bitsUsed < 1 || bitsUsed > 3 {
		bitsUsed = 1 // Default to 1 bit if invalid
	}

	// Get the file from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		sendErrorResponse(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input file path
	inputPath := filepath.Join(tempDir, "decode_"+timestamp+filepath.Ext(handler.Filename))

	// Save the uploaded file
	inputFile, err := os.Create(inputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	defer inputFile.Close()
	defer os.Remove(inputPath) // Clean up

	_, err = io.Copy(inputFile, file)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	inputFile.Close() // Close now so it can be read

	// Create LSB encoder
	encoder, err := steganography.NewLSBEncoder(seed, bitsUsed)
	if err != nil {
		sendErrorResponse(w, "Failed to create encoder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode the message
	message, err := encoder.DecodeMessage(inputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to decode message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the response
	sendSuccessResponse(w, "Message decoded successfully", map[string]string{
		"message": message,
	})
}

// Helper functions for API responses
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: false,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

func sendSuccessResponse(w http.ResponseWriter, message string, data any) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}
