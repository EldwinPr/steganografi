package steganography

import (
	"encoding/binary"
	"errors"
	"io"
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
	// Read AVI file
	header, videoData, err := readAviFile(inputPath)
	if err != nil {
		return err
	}

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

	// Write modified AVI file
	return writeAviFile(outputPath, header, videoData)
}

// DecodeMessage extracts a hidden text message from an AVI file
func (e *VideoEncoder) DecodeMessage(inputPath string) (string, error) {
	// Read AVI file
	_, videoData, err := readAviFile(inputPath)
	if err != nil {
		return "", err
	}

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

// readAviFile reads an AVI file and returns header and video data
func readAviFile(path string) ([]byte, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Read file size
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	fileSize := fileInfo.Size()

	// Check if file is too small to be a valid AVI
	if fileSize < 12 {
		return nil, nil, errors.New("file too small to be a valid AVI")
	}

	// Read full file content
	fileContent := make([]byte, fileSize)
	_, err = io.ReadFull(f, fileContent)
	if err != nil {
		return nil, nil, err
	}

	// Verify RIFF header
	if string(fileContent[0:4]) != "RIFF" || string(fileContent[8:12]) != "AVI " {
		return nil, nil, errors.New("not a valid AVI file")
	}

	// Find the movi chunk
	var moviOffset int64 = 0
	var moviSize int64 = 0

	// Navigate through chunks
	offset := int64(12) // Start after RIFF/AVI header

	for offset < fileSize-8 { // Need at least 8 bytes for a chunk header
		// Read chunk ID and size
		chunkID := string(fileContent[offset : offset+4])
		chunkSize := int64(binary.LittleEndian.Uint32(fileContent[offset+4 : offset+8]))

		// Make chunkSize even if it's odd (AVI padding rule)
		if chunkSize%2 == 1 {
			chunkSize++
		}

		// Check if this is LIST chunk
		if chunkID == "LIST" && offset+12 <= fileSize {
			// Check if this LIST contains movi
			listType := string(fileContent[offset+8 : offset+12])
			if listType == "movi" {
				moviOffset = offset + 12 // Skip LIST header and type
				moviSize = chunkSize - 4 // Subtract the size of the LIST type
				break
			}
		}

		// Move to next chunk
		offset += 8 + chunkSize

		// Safety check
		if offset >= fileSize {
			break
		}
	}

	if moviOffset == 0 || moviSize == 0 {
		return nil, nil, errors.New("movi chunk not found")
	}

	// Limit movi size to avoid memory issues
	if moviSize > fileSize-moviOffset {
		moviSize = fileSize - moviOffset
	}

	// Extract header and video data
	header := make([]byte, moviOffset)
	copy(header, fileContent[:moviOffset])

	videoData := make([]byte, moviSize)
	if moviOffset+moviSize > int64(len(fileContent)) {
		moviSize = int64(len(fileContent)) - moviOffset
	}
	copy(videoData, fileContent[moviOffset:moviOffset+moviSize])

	return header, videoData, nil
}

// writeAviFile writes header and video data to an AVI file
func writeAviFile(path string, header []byte, videoData []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header (which should include the LIST movi chunk header)
	_, err = f.Write(header)
	if err != nil {
		return err
	}

	// Write video data
	_, err = f.Write(videoData)
	if err != nil {
		return err
	}

	// Update the RIFF chunk size in the file header
	totalSize := len(header) + len(videoData)
	riffSize := uint32(totalSize - 8) // Total size minus the RIFF header and size fields

	// Go back to position 4 and write the updated size
	_, err = f.Seek(4, io.SeekStart)
	if err != nil {
		return err
	}

	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, riffSize)
	_, err = f.Write(sizeBytes)
	if err != nil {
		return err
	}

	// Update the LIST movi chunk size if we can find it
	// Search for LIST movi in the header
	for i := 0; i < len(header)-12; i++ {
		if string(header[i:i+4]) == "LIST" && string(header[i+8:i+12]) == "movi" {
			// Go to the size field position
			_, err = f.Seek(int64(i+4), io.SeekStart)
			if err != nil {
				return err
			}

			// Write the movi chunk size (plus 4 for the 'movi' identifier)
			moviSize := uint32(len(videoData) + 4)
			binary.LittleEndian.PutUint32(sizeBytes, moviSize)
			_, err = f.Write(sizeBytes)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}
