// audio.go - Audio steganography implementation
package steganography

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"os"
	"strconv"
)

// AudioEncoder handles LSB steganography for WAV files
type AudioEncoder struct {
	Seed int64
}

// NewAudioEncoder creates a new audio steganography encoder with the given seed
func NewAudioEncoder(seed string) (*AudioEncoder, error) {
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

	return &AudioEncoder{
		Seed: seedInt,
	}, nil
}

// EncodeData embeds binary data into a WAV file using LSB steganography
func (e *AudioEncoder) EncodeData(inputPath, outputPath string, data []byte) error {
	// Read WAV file
	header, audioData, err := readWavFile(inputPath)
	if err != nil {
		return err
	}

	// Calculate capacity (1 bit per sample)
	capacityBits := len(audioData)
	messageBits := len(data)*8 + 32 // 32 bits for length

	if messageBits > capacityBits {
		return errors.New("message exceeds audio capacity")
	}

	// Create full data with length prefix
	fullData := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(fullData[0:4], uint32(len(data)))
	copy(fullData[4:], data)

	// Generate sample indices based on seed
	indices := generateSampleOrder(len(audioData), len(fullData)*8, e.Seed)

	// Embed data
	bitIndex := 0
	for i := 0; i < len(fullData); i++ {
		byteVal := fullData[i]
		for b := 0; b < 8; b++ {
			bit := (byteVal >> (7 - b)) & 1
			sampleIndex := indices[bitIndex]
			// Clear LSB and set to message bit
			audioData[sampleIndex] = (audioData[sampleIndex] & 0xFE) | bit
			bitIndex++
		}
	}

	// Write modified WAV file
	return writeWavFile(outputPath, header, audioData)
}

// DecodeData extracts hidden binary data from a WAV file
func (e *AudioEncoder) DecodeData(inputPath string) ([]byte, error) {
	// Read WAV file
	_, audioData, err := readWavFile(inputPath)
	if err != nil {
		return nil, err
	}

	// Generate sample indices based on seed
	indices := generateSampleOrder(len(audioData), 32, e.Seed) // Start with enough for length

	// Extract length first
	var lengthBytes [4]byte
	for i := 0; i < 32; i++ {
		bit := audioData[indices[i]] & 1
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)
		lengthBytes[byteIndex] |= bit << bitPosition
	}

	dataLength := binary.BigEndian.Uint32(lengthBytes[:])
	if dataLength > uint32(len(audioData)/8) {
		return nil, errors.New("invalid data length")
	}

	// Generate indices for the full message
	indices = generateSampleOrder(len(audioData), int(dataLength)*8+32, e.Seed)

	// Extract data
	extractedData := make([]byte, dataLength)
	for i := 0; i < int(dataLength); i++ {
		for b := 0; b < 8; b++ {
			bitIndex := 32 + i*8 + b
			bit := audioData[indices[bitIndex]] & 1
			extractedData[i] |= bit << (7 - b)
		}
	}

	return extractedData, nil
}

// EncodeMessage is a convenience method that encodes a text message
func (e *AudioEncoder) EncodeMessage(inputPath, outputPath, message string) error {
	return e.EncodeData(inputPath, outputPath, []byte(message))
}

// DecodeMessage is a convenience method that decodes a text message
func (e *AudioEncoder) DecodeMessage(inputPath string) (string, error) {
	data, err := e.DecodeData(inputPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper functions

// generateSampleOrder creates a deterministic order of sample indices
func generateSampleOrder(totalSamples, requiredBits int, seed int64) []int {
	indices := make([]int, requiredBits)

	if seed < 0 {
		// Sequential mode
		for i := 0; i < requiredBits; i++ {
			indices[i] = i % totalSamples
		}
		return indices
	}

	// Random mode with seed
	rng := rand.New(rand.NewSource(seed))
	used := make(map[int]bool)

	for i := 0; i < requiredBits; {
		idx := rng.Intn(totalSamples)
		if !used[idx] {
			used[idx] = true
			indices[i] = idx
			i++
		}
	}

	return indices
}

// readWavFile reads a WAV file and returns header and audio data
// Improved readWavFile function
func readWavFile(path string) ([]byte, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Read RIFF header
	var riffID [4]byte
	var fileSize uint32
	var waveID [4]byte

	if err := binary.Read(f, binary.LittleEndian, &riffID); err != nil {
		return nil, nil, err
	}
	if string(riffID[:]) != "RIFF" {
		return nil, nil, errors.New("not a valid WAV file")
	}

	if err := binary.Read(f, binary.LittleEndian, &fileSize); err != nil {
		return nil, nil, err
	}

	if err := binary.Read(f, binary.LittleEndian, &waveID); err != nil {
		return nil, nil, err
	}
	if string(waveID[:]) != "WAVE" {
		return nil, nil, errors.New("not a valid WAV file")
	}

	// Find the data chunk
	var dataFound bool
	var dataSize uint32
	var totalHeaderSize int64 = 12 // RIFF + size + WAVE

	for !dataFound {
		var chunkID [4]byte
		var chunkSize uint32

		if err := binary.Read(f, binary.LittleEndian, &chunkID); err != nil {
			return nil, nil, err
		}
		if err := binary.Read(f, binary.LittleEndian, &chunkSize); err != nil {
			return nil, nil, err
		}

		totalHeaderSize += 8 // chunk ID + chunk size

		if string(chunkID[:]) == "data" {
			dataFound = true
			dataSize = chunkSize
		} else {
			// Skip this chunk
			totalHeaderSize += int64(chunkSize)
			if _, err := f.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, nil, err
			}
		}
	}

	// Go back to beginning and read entire header
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}

	header := make([]byte, totalHeaderSize)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, nil, err
	}

	// Read audio data
	audioData := make([]byte, dataSize)
	if _, err := io.ReadFull(f, audioData); err != nil {
		return nil, nil, err
	}

	return header, audioData, nil
}

// writeWavFile writes header and audio data to a WAV file
func writeWavFile(path string, header []byte, audioData []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header
	_, err = f.Write(header)
	if err != nil {
		return err
	}

	// Write data
	_, err = f.Write(audioData)
	if err != nil {
		return err
	}

	return nil
}
