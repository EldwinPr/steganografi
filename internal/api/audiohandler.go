package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"steganografi/internal/steganography"
)

// HandleAudioEncodeText handles the encoding of a text message into a WAV file
func HandleAudioEncodeText(w http.ResponseWriter, r *http.Request) {
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

	// Get the audio file from the form
	file, handler, err := r.FormFile("audio")
	if err != nil {
		sendErrorResponse(w, "Failed to get audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(handler.Filename)
	if ext != ".wav" {
		sendErrorResponse(w, "Only WAV files are supported", http.StatusBadRequest)
		return
	}

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input and output file paths
	inputPath := filepath.Join(tempDir, "input_"+timestamp+ext)
	outputPath := filepath.Join(tempDir, "output_"+timestamp+".wav")

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

	// Create audio encoder
	encoder, err := steganography.NewAudioEncoder(seed)
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
	w.Header().Set("Content-Disposition", "attachment; filename=stego_audio.wav")
	w.Header().Set("Content-Type", "audio/wav")

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

// HandleAudioDecodeText handles the decoding of a text message from a WAV file
func HandleAudioDecodeText(w http.ResponseWriter, r *http.Request) {
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

	// Get the audio file from the form
	file, handler, err := r.FormFile("audio")
	if err != nil {
		sendErrorResponse(w, "Failed to get audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(handler.Filename)
	if ext != ".wav" {
		sendErrorResponse(w, "Only WAV files are supported", http.StatusBadRequest)
		return
	}

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input file path
	inputPath := filepath.Join(tempDir, "decode_"+timestamp+ext)

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

	// Create audio encoder
	encoder, err := steganography.NewAudioEncoder(seed)
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
