// common.go

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Response represents the API response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
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

// Utility function to optimize image size
func optimizeImageSize(inputPath, outputPath string) error {
	// Read the original file to get its size
	originalFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer originalFile.Close()

	originalInfo, err := originalFile.Stat()
	if err != nil {
		return err
	}
	originalSize := originalInfo.Size()

	// Read the encoded file to get its size
	encodedFile, err := os.Open(outputPath)
	if err != nil {
		return err
	}
	defer encodedFile.Close()

	encodedInfo, err := encodedFile.Stat()
	if err != nil {
		return err
	}
	encodedSize := encodedInfo.Size()

	// If the encoded file is significantly larger, try to optimize it
	if float64(encodedSize) > float64(originalSize)*1.2 { // If more than 20% larger
		// For now, we'll just log this - in a real implementation,
		// you might use an image optimization library here
		fmt.Printf("Warning: Encoded file is %.2f%% larger than original\n",
			float64(encodedSize-originalSize)/float64(originalSize)*100)
	}

	return nil
}
