package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
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

// HandleEncodeText handles the encoding of a text message into an image
func HandleEncodeText(w http.ResponseWriter, r *http.Request) {
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

// HandleEncodeFile handles the encoding of a file into an image
func HandleEncodeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50 MB max
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

	// Get the carrier image file
	imageFile, imageHandler, err := r.FormFile("image")
	if err != nil {
		sendErrorResponse(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer imageFile.Close()

	// Get the file to hide
	dataFile, dataHandler, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "Failed to get data file", http.StatusBadRequest)
		return
	}
	defer dataFile.Close()

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input and output file paths
	inputImagePath := filepath.Join(tempDir, "input_image_"+timestamp+filepath.Ext(imageHandler.Filename))
	inputDataPath := filepath.Join(tempDir, "input_data_"+timestamp+filepath.Ext(dataHandler.Filename))
	outputPath := filepath.Join(tempDir, "output_"+timestamp+".png")

	// Save the uploaded image file
	inputImageFile, err := os.Create(inputImagePath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded image", http.StatusInternalServerError)
		return
	}
	defer inputImageFile.Close()
	defer os.Remove(inputImagePath) // Clean up

	_, err = io.Copy(inputImageFile, imageFile)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded image", http.StatusInternalServerError)
		return
	}
	inputImageFile.Close() // Close now so it can be read

	// Save the uploaded data file
	inputDataFile, err := os.Create(inputDataPath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded data file", http.StatusInternalServerError)
		return
	}
	defer inputDataFile.Close()
	defer os.Remove(inputDataPath) // Clean up

	_, err = io.Copy(inputDataFile, dataFile)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded data file", http.StatusInternalServerError)
		return
	}
	inputDataFile.Close() // Close now so it can be read

	// Read the data file
	fileData, err := ioutil.ReadFile(inputDataPath)
	if err != nil {
		sendErrorResponse(w, "Failed to read data file", http.StatusInternalServerError)
		return
	}

	// Prepare file metadata
	fileExt := filepath.Ext(dataHandler.Filename)
	fileName := dataHandler.Filename

	// Create metadata structure
	metadata := struct {
		FileName string `json:"fileName"`
		FileExt  string `json:"fileExt"`
		FileSize int    `json:"fileSize"`
	}{
		FileName: fileName,
		FileExt:  fileExt,
		FileSize: len(fileData),
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		sendErrorResponse(w, "Failed to create file metadata", http.StatusInternalServerError)
		return
	}

	// Combine metadata and file data
	// Format: [4 bytes metadata length][metadata JSON][file data]
	metadataLen := uint32(len(metadataJSON))
	combinedData := make([]byte, 4+len(metadataJSON)+len(fileData))

	// Write metadata length
	combinedData[0] = byte(metadataLen >> 24)
	combinedData[1] = byte(metadataLen >> 16)
	combinedData[2] = byte(metadataLen >> 8)
	combinedData[3] = byte(metadataLen)

	// Write metadata and file data
	copy(combinedData[4:], metadataJSON)
	copy(combinedData[4+len(metadataJSON):], fileData)

	// Create LSB encoder
	encoder, err := steganography.NewLSBEncoder(seed, bitsUsed)
	if err != nil {
		sendErrorResponse(w, "Failed to create encoder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the data
	err = encoder.EncodeData(inputImagePath, outputPath, combinedData)
	if err != nil {
		sendErrorResponse(w, "Failed to encode file: "+err.Error(), http.StatusInternalServerError)
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

// HandleDecodeText handles the decoding of a text message from an image
func HandleDecodeText(w http.ResponseWriter, r *http.Request) {
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

// HandleDecodeFile handles the decoding of a file from an image
func HandleDecodeFile(w http.ResponseWriter, r *http.Request) {
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

	// Decode the data
	data, err := encoder.DecodeData(inputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to decode data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if we have enough data for metadata length
	if len(data) < 4 {
		sendErrorResponse(w, "Invalid data format: too short", http.StatusInternalServerError)
		return
	}

	// Extract metadata length
	metadataLen := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])

	// Check if we have enough data for metadata
	if len(data) < 4+int(metadataLen) {
		sendErrorResponse(w, "Invalid data format: metadata incomplete", http.StatusInternalServerError)
		return
	}

	// Extract metadata
	metadataJSON := data[4 : 4+metadataLen]

	// Parse metadata
	var metadata struct {
		FileName string `json:"fileName"`
		FileExt  string `json:"fileExt"`
		FileSize int    `json:"fileSize"`
	}

	err = json.Unmarshal(metadataJSON, &metadata)
	if err != nil {
		sendErrorResponse(w, "Failed to parse file metadata", http.StatusInternalServerError)
		return
	}

	// Extract file data
	fileData := data[4+metadataLen:]

	// Check if file data length matches expected size
	if len(fileData) != metadata.FileSize {
		sendErrorResponse(w, "File data size mismatch", http.StatusInternalServerError)
		return
	}

	// Encode file data as base64 for JSON response
	fileBase64 := base64.StdEncoding.EncodeToString(fileData)

	// Send the response
	sendSuccessResponse(w, "File decoded successfully", map[string]interface{}{
		"fileName": metadata.FileName,
		"fileExt":  metadata.FileExt,
		"fileSize": metadata.FileSize,
		"fileData": fileBase64,
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
