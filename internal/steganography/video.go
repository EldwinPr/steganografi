package steganography

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"os"
	"strconv"
)

// VideoEncoder handles LSB steganography for AVI files
type VideoEncoder struct {
	Seed int64
}

// NewVideoEncoder creates a new video steganography encoder with the given seed
func NewVideoEncoder(seed string) (*VideoEncoder, error) {
	var seedInt int64 = -1 // Default to -1 (sequential)

	if seed != "" {
		// Convert seed string to int64
		var err error
		seedInt, err = strconv.ParseInt(seed, 10, 64)
		if err != nil {
			// If not a number, use string hash as seed
			h := 0
			for i := 0; i < len(seed); i++ {
				h = 31*h + int(seed[i])
			}
			seedInt = int64(h)
		}
	}

	return &VideoEncoder{
		Seed: seedInt,
	}, nil
}

// EncodeMessage embeds a text message into an AVI file using LSB steganography
func (e *VideoEncoder) EncodeMessage(inputPath, outputPath, message string) error {
	// Read original file directly to avoid modifying header structure
	originalData, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	// Validate AVI file
	if len(originalData) < 12 || string(originalData[0:4]) != "RIFF" || string(originalData[8:12]) != "AVI " {
		return errors.New("not a valid AVI file")
	}

	// Find movi data chunk
	moviOffset, moviLength, err := findMoviChunk(originalData)
	if err != nil {
		return err
	}

	// Copy the original data
	outputData := make([]byte, len(originalData))
	copy(outputData, originalData)

	// Get the actual video data part (inside movi chunk)
	videoData := outputData[moviOffset : moviOffset+moviLength]

	// Convert message to bytes
	data := []byte(message)

	// Calculate capacity (1 bit per byte)
	capacityBits := len(videoData)
	messageBits := len(data)*8 + 32 // 32 bits for length

	if messageBits > capacityBits {
		return errors.New("message exceeds video capacity")
	}

	// Create full data with length prefix
	fullData := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(fullData[0:4], uint32(len(data)))
	copy(fullData[4:], data)

	// Generate pixel indices based on seed
	indices := generatePixelOrderVid(len(videoData), len(fullData)*8, e.Seed)

	// Embed data
	bitIndex := 0
	for i := 0; i < len(fullData); i++ {
		byteVal := fullData[i]
		for b := 0; b < 8; b++ {
			bit := (byteVal >> (7 - b)) & 1
			pixelIndex := indices[bitIndex]
			if pixelIndex >= len(videoData) {
				return errors.New("pixel index out of bounds")
			}
			// Clear LSB and set to message bit
			videoData[pixelIndex] = (videoData[pixelIndex] & 0xFE) | bit
			bitIndex++
		}
	}

	// Write the modified file
	return os.WriteFile(outputPath, outputData, 0644)
}

// DecodeMessage extracts a hidden text message from an AVI file
func (e *VideoEncoder) DecodeMessage(inputPath string) (string, error) {
	// Read the entire AVI file
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	// Validate AVI file
	if len(fileData) < 12 || string(fileData[0:4]) != "RIFF" || string(fileData[8:12]) != "AVI " {
		return "", errors.New("not a valid AVI file")
	}

	// Find movi data chunk
	moviOffset, moviLength, err := findMoviChunk(fileData)
	if err != nil {
		return "", err
	}

	// Get video data
	videoData := fileData[moviOffset : moviOffset+moviLength]

	// Generate pixel indices based on seed
	indices := generatePixelOrderVid(len(videoData), 32, e.Seed) // Start with enough for length

	// Validate indices
	for _, idx := range indices {
		if idx >= len(videoData) {
			return "", errors.New("pixel index out of bounds during decoding")
		}
	}

	// Extract length first
	var lengthBytes [4]byte
	for i := 0; i < 32; i++ {
		bit := videoData[indices[i]] & 1
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)
		lengthBytes[byteIndex] |= bit << bitPosition
	}

	dataLength := binary.BigEndian.Uint32(lengthBytes[:])
	if dataLength > uint32(len(videoData)/8) {
		return "", errors.New("invalid data length")
	}

	// Generate indices for the full message
	indices = generatePixelOrderVid(len(videoData), int(dataLength)*8+32, e.Seed)

	// Validate all indices
	for _, idx := range indices {
		if idx >= len(videoData) {
			return "", errors.New("pixel index out of bounds during data extraction")
		}
	}

	// Extract data
	extractedData := make([]byte, dataLength)
	for i := 0; i < int(dataLength); i++ {
		for b := 0; b < 8; b++ {
			bitIndex := 32 + i*8 + b
			bit := videoData[indices[bitIndex]] & 1
			extractedData[i] |= bit << (7 - b)
		}
	}

	return string(extractedData), nil
}

// Helper functions

// findMoviChunk locates the movi chunk in AVI data and returns its offset and length
func findMoviChunk(fileData []byte) (int, int, error) {
	fileSize := len(fileData)
	offset := 12 // Start after RIFF/AVI header

	for offset < fileSize-8 {
		// Read chunk ID and size
		chunkID := string(fileData[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(fileData[offset+4 : offset+8]))

		// Make chunkSize even if it's odd (AVI padding rule)
		if chunkSize%2 == 1 {
			chunkSize++
		}

		// Check if this is LIST chunk
		if chunkID == "LIST" && offset+12 <= fileSize {
			// Check if this LIST contains movi
			listType := string(fileData[offset+8 : offset+12])
			if listType == "movi" {
				// Found movi chunk
				dataOffset := offset + 12   // Skip LIST header and type
				dataLength := chunkSize - 4 // Subtract the size of the LIST type

				// Ensure we don't exceed the file bounds
				if dataOffset+dataLength > fileSize {
					dataLength = fileSize - dataOffset
				}

				return dataOffset, dataLength, nil
			}
		}

		// Move to next chunk
		offset += 8 + chunkSize

		// Safety check
		if offset >= fileSize {
			break
		}
	}

	return 0, 0, errors.New("movi chunk not found")
}

// generatePixelOrderVid creates a deterministic order of pixel indices
func generatePixelOrderVid(totalPixels, requiredBits int, seed int64) []int {
	indices := make([]int, requiredBits)

	if seed < 0 {
		// Sequential mode
		for i := 0; i < requiredBits; i++ {
			indices[i] = i % totalPixels
		}
		return indices
	}

	// Random mode with seed
	rng := rand.New(rand.NewSource(seed))
	used := make(map[int]bool)

	for i := 0; i < requiredBits; {
		idx := rng.Intn(totalPixels)
		if !used[idx] {
			used[idx] = true
			indices[i] = idx
			i++
		}
	}

	return indices
}
