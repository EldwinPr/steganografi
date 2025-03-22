// fileutils.go

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// FileMetadata represents the metadata of a file to be encoded/decoded
type FileMetadata struct {
	FileName string `json:"fileName"`
	FileExt  string `json:"fileExt"`
	FileSize int    `json:"fileSize"`
}

// PrepareTemporaryPaths creates temporary file paths for processing
func PrepareTemporaryPaths(handler *multipart.FileHeader, prefix string) (string, string, string) {
	tempDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	inputPath := filepath.Join(tempDir, prefix+"_input_"+timestamp+filepath.Ext(handler.Filename))
	outputPath := filepath.Join(tempDir, prefix+"_output_"+timestamp+".png")

	return tempDir, inputPath, outputPath
}

// SaveUploadedFile saves an uploaded file to the specified path
func SaveUploadedFile(file multipart.File, path string) error {
	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, file)
	return err
}

// PrepareFileData reads a file and prepares its metadata and combined data
func PrepareFileData(filePath string, fileName string) ([]byte, error) {
	// Read the data file
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Prepare file metadata
	fileExt := filepath.Ext(fileName)

	// Create metadata structure
	metadata := FileMetadata{
		FileName: fileName,
		FileExt:  fileExt,
		FileSize: len(fileData),
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
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

	return combinedData, nil
}

// ExtractFileData extracts file metadata and data from combined data
func ExtractFileData(data []byte) (FileMetadata, []byte, error) {
	var metadata FileMetadata

	// Check if we have enough data for metadata length
	if len(data) < 4 {
		return metadata, nil, fmt.Errorf("invalid data format: too short")
	}

	// Extract metadata length
	metadataLen := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])

	// Check if we have enough data for metadata
	if len(data) < 4+int(metadataLen) {
		return metadata, nil, fmt.Errorf("invalid data format: metadata incomplete")
	}

	// Extract metadata
	metadataJSON := data[4 : 4+metadataLen]

	// Parse metadata
	err := json.Unmarshal(metadataJSON, &metadata)
	if err != nil {
		return metadata, nil, fmt.Errorf("failed to parse file metadata: %v", err)
	}

	// Extract file data
	fileData := data[4+metadataLen:]

	// Check if file data length matches expected size
	if len(fileData) != metadata.FileSize {
		return metadata, nil, fmt.Errorf("file data size mismatch")
	}

	return metadata, fileData, nil
}

// SendFileForDownload sends a file as a download response
func SendFileForDownload(w http.ResponseWriter, filePath string, fileName string) error {
	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "image/png")

	// Send the file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}
