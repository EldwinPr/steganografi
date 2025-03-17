package steganography

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/big"
	mathrand "math/rand"
	"os"
	"strconv"
)

// LSBEncoder handles LSB steganography encoding
type LSBEncoder struct {
	Seed     int64
	BitsUsed int // Number of LSB bits to use (1-3 recommended)
}

// NewLSBEncoder creates a new LSB encoder with the given seed
func NewLSBEncoder(seed string, bitsUsed int) (*LSBEncoder, error) {
	if bitsUsed < 1 || bitsUsed > 3 {
		return nil, errors.New("bits used must be between 1 and 3")
	}

	var seedInt int64
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
	} else {
		// Generate random seed if none provided
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return nil, err
		}
		seedInt = n.Int64()
	}

	return &LSBEncoder{
		Seed:     seedInt,
		BitsUsed: bitsUsed,
	}, nil
}

// EncodeData embeds binary data into an image using LSB steganography
func (e *LSBEncoder) EncodeData(inputPath, outputPath string, data []byte) error {
	// Open the input image
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return err
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

	// Check if the data can fit in the image
	maxBytes := (width*height*3*e.BitsUsed)/8 - 8 // 8 bytes for length and file extension length
	if int(dataLength) > maxBytes {
		return errors.New("data too large for the image")
	}

	// Create a byte slice for the length (4 bytes) + data
	fullData := make([]byte, 4+dataLength)
	binary.BigEndian.PutUint32(fullData[0:4], dataLength)
	copy(fullData[4:], data)

	// Use seed to determine pixel order
	rng := NewSeededRNG(e.Seed)
	pixels := generatePixelOrder(width, height, rng)

	// Embed the data
	bitIndex := 0
	for _, pixel := range pixels {
		x, y := pixel.X, pixel.Y

		if bitIndex/8 >= len(fullData) {
			break
		}

		r, g, b, a := rgbaImg.At(x, y).RGBA()

		// Process each color channel
		if bitIndex/8 < len(fullData) {
			r = embedBits(r, fullData, bitIndex, e.BitsUsed)
			bitIndex += e.BitsUsed
		}

		if bitIndex/8 < len(fullData) {
			g = embedBits(g, fullData, bitIndex, e.BitsUsed)
			bitIndex += e.BitsUsed
		}

		if bitIndex/8 < len(fullData) {
			b = embedBits(b, fullData, bitIndex, e.BitsUsed)
			bitIndex += e.BitsUsed
		}

		// Set the modified pixel
		rgbaImg.Set(x, y, color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		})
	}

	// Save the output image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return png.Encode(outFile, rgbaImg)
}

// DecodeData extracts hidden binary data from an image
func (e *LSBEncoder) DecodeData(inputPath string) ([]byte, error) {
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

	// Get image bounds
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Use seed to determine pixel order
	rng := NewSeededRNG(e.Seed)
	pixels := generatePixelOrder(width, height, rng)

	// Extract the length first (4 bytes)
	extractedData := make([]byte, 0)
	bitIndex := 0
	currentByte := byte(0)
	bitsRead := 0

	for _, pixel := range pixels {
		x, y := pixel.X, pixel.Y
		r, g, b, _ := img.At(x, y).RGBA()

		// Extract from red channel
		extractBits(r, &currentByte, &bitsRead, &bitIndex, e.BitsUsed, &extractedData)
		if len(extractedData) >= 4 && len(extractedData) >= int(binary.BigEndian.Uint32(extractedData[0:4]))+4 {
			break
		}

		// Extract from green channel
		extractBits(g, &currentByte, &bitsRead, &bitIndex, e.BitsUsed, &extractedData)
		if len(extractedData) >= 4 && len(extractedData) >= int(binary.BigEndian.Uint32(extractedData[0:4]))+4 {
			break
		}

		// Extract from blue channel
		extractBits(b, &currentByte, &bitsRead, &bitIndex, e.BitsUsed, &extractedData)
		if len(extractedData) >= 4 && len(extractedData) >= int(binary.BigEndian.Uint32(extractedData[0:4]))+4 {
			break
		}
	}

	if len(extractedData) < 4 {
		return nil, errors.New("could not extract data length")
	}

	// Get the data length
	dataLength := binary.BigEndian.Uint32(extractedData[0:4])

	if len(extractedData) < int(dataLength)+4 {
		return nil, errors.New("extracted data is shorter than expected")
	}

	// Extract the data
	data := extractedData[4 : 4+dataLength]
	return data, nil
}

// EncodeMessage is a convenience method that encodes a text message
func (e *LSBEncoder) EncodeMessage(inputPath, outputPath, message string) error {
	return e.EncodeData(inputPath, outputPath, []byte(message))
}

// DecodeMessage is a convenience method that decodes a text message
func (e *LSBEncoder) DecodeMessage(inputPath string) (string, error) {
	data, err := e.DecodeData(inputPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper functions

// Pixel represents a coordinate in the image
type Pixel struct {
	X, Y int
}

// NewSeededRNG creates a deterministic random number generator from a seed
func NewSeededRNG(seed int64) *mathrand.Rand {
	source := mathrand.NewSource(seed)
	return mathrand.New(source)
}

// generatePixelOrder creates a pseudo-random order of pixels based on the seed
func generatePixelOrder(width, height int, rng *mathrand.Rand) []Pixel {
	pixels := make([]Pixel, width*height)

	// Initialize with all pixels
	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels[index] = Pixel{X: x, Y: y}
			index++
		}
	}

	// Shuffle the pixels using Fisher-Yates algorithm
	for i := len(pixels) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		pixels[i], pixels[j] = pixels[j], pixels[i]
	}

	return pixels
}

// embedBits embeds bits from data into the color value
func embedBits(colorValue uint32, data []byte, bitIndex int, bitsUsed int) uint32 {
	// Convert to 8-bit color
	colorValue = colorValue >> 8

	// Create a mask to clear the LSBs
	mask := uint32(0xFF) << bitsUsed
	colorValue = colorValue & mask

	// Calculate which bits to embed
	byteIndex := bitIndex / 8
	bitOffset := bitIndex % 8

	if byteIndex < len(data) {
		// Get the bits to embed
		bitsToEmbed := uint32(0)

		// Handle the case where we need bits from two bytes
		if bitOffset+bitsUsed <= 8 {
			// All bits come from the same byte
			bitsToEmbed = uint32((data[byteIndex] >> (8 - bitOffset - bitsUsed)) & ((1 << bitsUsed) - 1))
		} else {
			// Bits come from two consecutive bytes
			firstByteBits := 8 - bitOffset
			secondByteBits := bitsUsed - firstByteBits

			firstPart := uint32((data[byteIndex] & ((1 << firstByteBits) - 1)) << secondByteBits)
			secondPart := uint32(0)

			if byteIndex+1 < len(data) {
				secondPart = uint32(data[byteIndex+1] >> (8 - secondByteBits))
			}

			bitsToEmbed = firstPart | secondPart
		}

		// Embed the bits
		colorValue = colorValue | bitsToEmbed
	}

	// Convert back to 16-bit color
	return colorValue << 8
}

// extractBits extracts bits from the color value
func extractBits(colorValue uint32, currentByte *byte, bitsRead *int, bitIndex *int,
	bitsUsed int, extractedData *[]byte) {
	// Convert to 8-bit color
	colorValue = colorValue >> 8

	// Extract the LSBs
	extractedBits := uint8(colorValue & ((1 << bitsUsed) - 1))

	// Add the bits to the current byte
	*currentByte = (*currentByte << bitsUsed) | extractedBits
	*bitsRead += bitsUsed

	// If we've read 8 bits, add the byte to our data
	if *bitsRead >= 8 {
		*extractedData = append(*extractedData, *currentByte)
		*currentByte = 0
		*bitsRead = 0
	}

	*bitIndex += bitsUsed
}
