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

// HandleVideoEncodeText handles the encoding of a text message into an AVI file
func HandleVideoEncodeText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(30 << 20) // 30 MB max for video
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")
	message := r.FormValue("message")

	// Get the video file from the form
	file, handler, err := r.FormFile("video")
	if err != nil {
		sendErrorResponse(w, "Failed to get video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(handler.Filename)
	if ext != ".avi" {
		sendErrorResponse(w, "Only AVI files are supported", http.StatusBadRequest)
		return
	}

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create input and output file paths
	inputPath := filepath.Join(tempDir, "input_"+timestamp+ext)
	outputPath := filepath.Join(tempDir, "output_"+timestamp+".avi")

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

	// Create video encoder
	encoder, err := steganography.NewVideoEncoder(seed)
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
	w.Header().Set("Content-Disposition", "attachment; filename=stego_video.avi")
	w.Header().Set("Content-Type", "video/x-msvideo")

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

// HandleVideoDecodeText handles the decoding of a text message from an AVI file
func HandleVideoDecodeText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(30 << 20) // 30 MB max for video
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")

	// Get the video file from the form
	file, handler, err := r.FormFile("video")
	if err != nil {
		sendErrorResponse(w, "Failed to get video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(handler.Filename)
	if ext != ".avi" {
		sendErrorResponse(w, "Only AVI files are supported", http.StatusBadRequest)
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

	// Create video encoder
	encoder, err := steganography.NewVideoEncoder(seed)
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

// HandleVideoEncodeFile handles the encoding of a file into an AVI file
func HandleVideoEncodeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(30 << 20) // 30 MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")

	// Get the carrier video file
	videoFile, videoHandler, err := r.FormFile("video")
	if err != nil {
		sendErrorResponse(w, "Failed to get video file", http.StatusBadRequest)
		return
	}
	defer videoFile.Close()

	// Get the data file to encode
	dataFile, dataHandler, err := r.FormFile("dataFile")
	if err != nil {
		sendErrorResponse(w, "Failed to get data file", http.StatusBadRequest)
		return
	}
	defer dataFile.Close()

	// Validate file extension
	videoExt := filepath.Ext(videoHandler.Filename)
	if videoExt != ".avi" {
		sendErrorResponse(w, "Only AVI files are supported as carriers", http.StatusBadRequest)
		return
	}

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create paths
	videoPath := filepath.Join(tempDir, "input_video_"+timestamp+videoExt)
	dataPath := filepath.Join(tempDir, "input_data_"+timestamp+filepath.Ext(dataHandler.Filename))
	outputPath := filepath.Join(tempDir, "output_"+timestamp+".avi")

	// Save the uploaded video file
	videoInputFile, err := os.Create(videoPath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded video", http.StatusInternalServerError)
		return
	}
	defer videoInputFile.Close()
	defer os.Remove(videoPath) // Clean up

	_, err = io.Copy(videoInputFile, videoFile)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded video", http.StatusInternalServerError)
		return
	}
	videoInputFile.Close() // Close now so it can be read

	// Save the uploaded data file
	dataInputFile, err := os.Create(dataPath)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded data file", http.StatusInternalServerError)
		return
	}
	defer dataInputFile.Close()
	defer os.Remove(dataPath) // Clean up

	_, err = io.Copy(dataInputFile, dataFile)
	if err != nil {
		sendErrorResponse(w, "Failed to save uploaded data file", http.StatusInternalServerError)
		return
	}
	dataInputFile.Close() // Close now so it can be read

	// Prepare file data with metadata
	fileData, err := PrepareFileData(dataPath, dataHandler.Filename)
	if err != nil {
		sendErrorResponse(w, "Failed to prepare file data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create video encoder
	encoder, err := steganography.NewVideoEncoder(seed)
	if err != nil {
		sendErrorResponse(w, "Failed to create encoder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the file data
	err = encoder.EncodeData(videoPath, outputPath, fileData)
	if err != nil {
		sendErrorResponse(w, "Failed to encode file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer os.Remove(outputPath) // Clean up

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=stego_video.avi")
	w.Header().Set("Content-Type", "video/x-msvideo")

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

// HandleVideoDecodeFile handles the decoding of a file from an AVI file
func HandleVideoDecodeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(30 << 20) // 30 MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	seed := r.FormValue("seed")

	// Get the video file from the form
	file, handler, err := r.FormFile("video")
	if err != nil {
		sendErrorResponse(w, "Failed to get video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(handler.Filename)
	if ext != ".avi" {
		sendErrorResponse(w, "Only AVI files are supported", http.StatusBadRequest)
		return
	}

	// Create a temporary directory for processing
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create paths
	inputPath := filepath.Join(tempDir, "decode_"+timestamp+ext)
	outputDir := filepath.Join(tempDir, "output_"+timestamp)

	// Create output directory
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		sendErrorResponse(w, "Failed to create output directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(outputDir) // Clean up

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

	// Create video encoder
	encoder, err := steganography.NewVideoEncoder(seed)
	if err != nil {
		sendErrorResponse(w, "Failed to create encoder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode the file data
	extractedData, err := encoder.DecodeData(inputPath)
	if err != nil {
		sendErrorResponse(w, "Failed to decode data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract file metadata and data
	metadata, fileData, err := ExtractFileData(extractedData)
	if err != nil {
		sendErrorResponse(w, "Failed to extract file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create output file path
	outputFilePath := filepath.Join(outputDir, metadata.FileName)

	// Save the extracted file
	err = os.WriteFile(outputFilePath, fileData, 0644)
	if err != nil {
		sendErrorResponse(w, "Failed to save extracted file", http.StatusInternalServerError)
		return
	}

	// Set appropriate content type based on file extension
	contentType := "application/octet-stream"
	switch metadata.FileExt {
	case ".pdf":
		contentType = "application/pdf"
	case ".txt":
		contentType = "text/plain"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename="+metadata.FileName)
	w.Header().Set("Content-Type", contentType)

	// Send the file
	outputFile, err := os.Open(outputFilePath)
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
