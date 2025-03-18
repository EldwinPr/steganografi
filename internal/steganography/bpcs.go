package steganography

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BPCSEncoder handles BPCS steganography encoding
type BPCSEncoder struct {
	Seed                int64
	ComplexityThreshold float64 // Threshold for determining complex regions (0.3-0.5 recommended)
}

// NewBPCSEncoder creates a new BPCS encoder with the given seed
func NewBPCSEncoder(seed string, complexityThreshold float64) (*BPCSEncoder, error) {
	var seedInt int64 = -1 // Default to -1

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

	// Validate complexity threshold
	if complexityThreshold < 0.3 || complexityThreshold > 0.5 {
		complexityThreshold = 0.45 // Default value if out of range
	}

	return &BPCSEncoder{
		Seed:                seedInt,
		ComplexityThreshold: complexityThreshold,
	}, nil
}

// EncodeData embeds binary data into an image using BPCS steganography
func (e *BPCSEncoder) EncodeData(inputPath, outputPath string, data []byte) error {
	// Check if input is JPG and convert if needed
	ext := strings.ToLower(filepath.Ext(inputPath))
	var img image.Image
	var err error

	// Open the input image
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the image based on format
	if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
		if err != nil {
			return err
		}
	} else {
		img, _, err = image.Decode(file)
		if err != nil {
			return err
		}
	}

	// Get image bounds
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create a new RGBA image to modify
	rgbaImg := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	// Get data length
	dataLength := uint32(len(data))

	// Create a byte slice for the length (4 bytes) + data
	fullData := make([]byte, 4+dataLength)
	binary.BigEndian.PutUint32(fullData[0:4], dataLength)
	copy(fullData[4:], data)

	// Convert data to bit planes
	dataBlocks := convertDataToBlocks(fullData)

	// Conjugate blocks to ensure complexity
	for i := range dataBlocks {
		if calculateComplexity(dataBlocks[i]) < e.ComplexityThreshold {
			dataBlocks[i] = conjugateBlock(dataBlocks[i])
			// Mark this block as conjugated (set a flag in the header)
			// This would be implemented in a real system
		}
	}

	// Find complex regions in the image and embed data
	err = e.embedDataInComplexRegions(rgbaImg, dataBlocks)
	if err != nil {
		return err
	}

	// Save the output image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Use no compression for PNG to minimize file size changes
	encoder := &png.Encoder{
		CompressionLevel: png.NoCompression,
	}

	return encoder.Encode(outFile, rgbaImg)
}

// DecodeData extracts hidden binary data from an image
func (e *BPCSEncoder) DecodeData(inputPath string) ([]byte, error) {
	// Open the input image
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Extract data blocks from complex regions
	dataBlocks, err := e.extractDataFromComplexRegions(img)
	if err != nil {
		return nil, err
	}

	// Deconjugate blocks if needed
	for i := range dataBlocks {
		// Check if this block was conjugated (would check a flag in the header)
		// For now, we'll check complexity and deconjugate if needed
		if calculateComplexity(dataBlocks[i]) > 0.9 {
			dataBlocks[i] = conjugateBlock(dataBlocks[i])
		}
	}

	// Convert blocks back to bytes
	fullData := convertBlocksToData(dataBlocks)

	if len(fullData) < 4 {
		return nil, errors.New("could not extract data length")
	}

	// Get the data length
	dataLength := binary.BigEndian.Uint32(fullData[0:4])

	if len(fullData) < int(dataLength)+4 {
		return nil, errors.New("extracted data is shorter than expected")
	}

	// Extract the data
	data := fullData[4 : 4+dataLength]
	return data, nil
}

// EncodeMessage is a convenience method that encodes a text message
func (e *BPCSEncoder) EncodeMessage(inputPath, outputPath, message string) error {
	return e.EncodeData(inputPath, outputPath, []byte(message))
}

// DecodeMessage is a convenience method that decodes a text message
func (e *BPCSEncoder) DecodeMessage(inputPath string) (string, error) {
	data, err := e.DecodeData(inputPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper functions for BPCS

// Block represents an 8x8 bit block
type Block [8][8]bool

// calculateComplexity calculates the complexity of a bit block
// Complexity is defined as the number of bit transitions / maximum possible transitions
func calculateComplexity(block Block) float64 {
	transitions := 0
	maxTransitions := 2 * 8 * 7 // Maximum possible transitions in an 8x8 block

	// Count horizontal transitions
	for i := 0; i < 8; i++ {
		for j := 0; j < 7; j++ {
			if block[i][j] != block[i][j+1] {
				transitions++
			}
		}
	}

	// Count vertical transitions
	for j := 0; j < 8; j++ {
		for i := 0; i < 7; i++ {
			if block[i][j] != block[i+1][j] {
				transitions++
			}
		}
	}

	return float64(transitions) / float64(maxTransitions)
}

// conjugateBlock performs the conjugation operation on a block
// This is used to ensure blocks have high complexity
func conjugateBlock(block Block) Block {
	var result Block

	// XOR with checkerboard pattern
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			result[i][j] = block[i][j] != ((i+j)%2 == 0)
		}
	}

	return result
}

// convertDataToBlocks converts a byte array to bit blocks
func convertDataToBlocks(data []byte) []Block {
	// Calculate how many blocks we need
	blockCount := (len(data)*8 + 63) / 64 // Each block holds 64 bits (8x8)
	blocks := make([]Block, blockCount)

	for byteIndex := 0; byteIndex < len(data); byteIndex++ {
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			// Calculate which block and position this bit belongs to
			blockIndex := (byteIndex*8 + bitIndex) / 64
			if blockIndex >= len(blocks) {
				break
			}

			position := (byteIndex*8 + bitIndex) % 64
			row := position / 8
			col := position % 8

			// Set the bit in the block
			bit := (data[byteIndex] >> (7 - bitIndex)) & 1
			blocks[blockIndex][row][col] = bit == 1
		}
	}

	return blocks
}

// convertBlocksToData converts bit blocks back to a byte array
func convertBlocksToData(blocks []Block) []byte {
	// Calculate how many bytes we need
	byteCount := (len(blocks)*64 + 7) / 8 // Each block holds 64 bits
	data := make([]byte, byteCount)

	for blockIndex, block := range blocks {
		for row := 0; row < 8; row++ {
			for col := 0; col < 8; col++ {
				// Calculate which byte and bit position this belongs to
				bitPosition := blockIndex*64 + row*8 + col
				byteIndex := bitPosition / 8
				if byteIndex >= len(data) {
					break
				}

				bitIndex := 7 - (bitPosition % 8)

				// Set the bit in the byte
				if block[row][col] {
					data[byteIndex] |= 1 << bitIndex
				}
			}
		}
	}

	return data
}

// embedDataInComplexRegions embeds data blocks in complex regions of the image
func (e *BPCSEncoder) embedDataInComplexRegions(img *image.RGBA, dataBlocks []Block) error {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Calculate how many 8x8 blocks we can fit in the image
	blockCountX := width / 8
	blockCountY := height / 8

	// Use seed to determine block order
	rng := NewSeededRNG(e.Seed)
	blockOrder := generateBlockOrder(blockCountX, blockCountY, rng)

	// Track which data block we're currently embedding
	currentDataBlock := 0

	// For each bit plane (0-7) in each color channel (R,G,B)
	for plane := 0; plane < 8; plane++ {
		// Skip the most significant bit planes to preserve image quality
		if plane > 5 {
			continue
		}

		for _, blockPos := range blockOrder {
			if currentDataBlock >= len(dataBlocks) {
				return nil // All data has been embedded
			}

			// Extract the current image block for each channel
			redBlock := extractBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 0)
			greenBlock := extractBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 1)
			blueBlock := extractBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 2)

			// Check if the blocks are complex enough to hide data
			if calculateComplexity(redBlock) > e.ComplexityThreshold {
				// Replace with data block
				embedBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 0, dataBlocks[currentDataBlock])
				currentDataBlock++

				if currentDataBlock >= len(dataBlocks) {
					return nil // All data has been embedded
				}
			}

			if calculateComplexity(greenBlock) > e.ComplexityThreshold {
				// Replace with data block
				embedBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 1, dataBlocks[currentDataBlock])
				currentDataBlock++

				if currentDataBlock >= len(dataBlocks) {
					return nil // All data has been embedded
				}
			}

			if calculateComplexity(blueBlock) > e.ComplexityThreshold {
				// Replace with data block
				embedBitPlaneBlock(img, blockPos.X, blockPos.Y, plane, 2, dataBlocks[currentDataBlock])
				currentDataBlock++

				if currentDataBlock >= len(dataBlocks) {
					return nil // All data has been embedded
				}
			}
		}
	}

	if currentDataBlock < len(dataBlocks) {
		return errors.New("not enough complex regions to embed all data")
	}

	return nil
}

// extractDataFromComplexRegions extracts data blocks from complex regions
func (e *BPCSEncoder) extractDataFromComplexRegions(img image.Image) ([]Block, error) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Calculate how many 8x8 blocks we can fit in the image
	blockCountX := width / 8
	blockCountY := height / 8

	// Use seed to determine block order
	rng := NewSeededRNG(e.Seed)
	blockOrder := generateBlockOrder(blockCountX, blockCountY, rng)

	// Store extracted data blocks
	var dataBlocks []Block

	// Create a temporary RGBA image for easier pixel manipulation
	rgbaImg := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	// For each bit plane (0-7) in each color channel (R,G,B)
	for plane := 0; plane < 8; plane++ {
		// Skip the most significant bit planes where we didn't hide data
		if plane > 5 {
			continue
		}

		for _, blockPos := range blockOrder {
			// Extract the current image block for each channel
			redBlock := extractBitPlaneBlock(rgbaImg, blockPos.X, blockPos.Y, plane, 0)
			greenBlock := extractBitPlaneBlock(rgbaImg, blockPos.X, blockPos.Y, plane, 1)
			blueBlock := extractBitPlaneBlock(rgbaImg, blockPos.X, blockPos.Y, plane, 2)

			// Check if the blocks are complex enough to contain data
			if calculateComplexity(redBlock) > e.ComplexityThreshold {
				dataBlocks = append(dataBlocks, redBlock)
			}

			if calculateComplexity(greenBlock) > e.ComplexityThreshold {
				dataBlocks = append(dataBlocks, greenBlock)
			}

			if calculateComplexity(blueBlock) > e.ComplexityThreshold {
				dataBlocks = append(dataBlocks, blueBlock)
			}
		}
	}

	if len(dataBlocks) == 0 {
		return nil, errors.New("no data blocks found")
	}

	return dataBlocks, nil
}

// BlockPosition represents the position of an 8x8 block in the image
type BlockPosition struct {
	X, Y int // Top-left corner of the block
}

// generateBlockOrder creates a pseudo-random order of blocks based on the seed
func generateBlockOrder(blockCountX, blockCountY int, rng *mathrand.Rand) []BlockPosition {
	blocks := make([]BlockPosition, blockCountX*blockCountY)

	// Initialize with all blocks
	index := 0
	for y := 0; y < blockCountY; y++ {
		for x := 0; x < blockCountX; x++ {
			blocks[index] = BlockPosition{X: x * 8, Y: y * 8}
			index++
		}
	}

	// Shuffle the blocks using Fisher-Yates algorithm
	for i := len(blocks) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}

	return blocks
}

// extractBitPlaneBlock extracts an 8x8 block from a specific bit plane of a color channel
func extractBitPlaneBlock(img *image.RGBA, startX, startY, plane, channel int) Block {
	var block Block

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			// Get the pixel value
			r, g, b, _ := img.At(startX+x, startY+y).RGBA()

			// Select the appropriate channel
			var value uint32
			switch channel {
			case 0:
				value = r >> 8 // Convert from 16-bit to 8-bit
			case 1:
				value = g >> 8
			case 2:
				value = b >> 8
			}

			// Extract the bit from the specified plane
			bit := (value >> uint(plane)) & 1
			block[y][x] = bit == 1
		}
	}

	return block
}

// embedBitPlaneBlock embeds an 8x8 block into a specific bit plane of a color channel
func embedBitPlaneBlock(img *image.RGBA, startX, startY, plane, channel int, block Block) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			// Get the pixel value
			r, g, b, a := img.At(startX+x, startY+y).RGBA()

			// Convert to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Create a mask to clear the bit at the specified plane
			mask := uint8(^(1 << uint(plane)))

			// Clear the bit and set it according to the block
			switch channel {
			case 0:
				r8 = (r8 & mask)
				if block[y][x] {
					r8 |= (1 << uint(plane))
				}
			case 1:
				g8 = (g8 & mask)
				if block[y][x] {
					g8 |= (1 << uint(plane))
				}
			case 2:
				b8 = (b8 & mask)
				if block[y][x] {
					b8 |= (1 << uint(plane))
				}
			}

			// Update the pixel
			img.Set(startX+x, startY+y, color.RGBA{r8, g8, b8, a8})
		}
	}
}
