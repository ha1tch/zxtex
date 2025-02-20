package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	_ "image/gif"
	_ "image/bmp"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// ZX Spectrum palette: 16 colors.
var ZXPalette = []color.RGBA{
	{0, 0, 0, 255},       // 0: Black
	{0, 0, 215, 255},     // 1: Blue
	{215, 0, 0, 255},     // 2: Red
	{215, 0, 215, 255},   // 3: Magenta
	{0, 215, 0, 255},     // 4: Green
	{0, 215, 215, 255},   // 5: Cyan
	{215, 215, 0, 255},   // 6: Yellow
	{215, 215, 215, 255}, // 7: White (normal)
	{0, 0, 0, 255},       // 8: Bright Black (same as black)
	{0, 0, 255, 255},     // 9: Bright Blue
	{255, 0, 0, 255},     // A: Bright Red
	{255, 0, 255, 255},   // B: Bright Magenta
	{0, 255, 0, 255},     // C: Bright Green
	{0, 255, 255, 255},   // D: Bright Cyan
	{255, 255, 0, 255},   // E: Bright Yellow
	{255, 255, 255, 255}, // F: Bright White
}

// nearestColor returns the index of the nearest ZX Spectrum palette color
// for the given color.
func nearestColor(r, g, b uint32) int {
	bestIndex := 0
	bestDist := math.MaxFloat64
	// Convert 16-bit colors to 8-bit.
	cr := float64(r >> 8)
	cg := float64(g >> 8)
	cb := float64(b >> 8)
	for i, pal := range ZXPalette {
		dr := cr - float64(pal.R)
		dg := cg - float64(pal.G)
		db := cb - float64(pal.B)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			bestIndex = i
		}
	}
	return bestIndex
}

// imageToHex converts an image file into a hex string representing each pixel.
func imageToHex(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Decode the image.
	img, format, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	_ = format

	// Ensure the image is in RGBA format.
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	var sb strings.Builder
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := rgba.At(x, y)
			r, g, b, _ := c.RGBA()
			idx := nearestColor(r, g, b)
			sb.WriteString(strings.ToUpper(strconv.FormatInt(int64(idx), 16)))
		}
	}
	return sb.String(), nil
}

// filterHexString removes all characters from the input string that are not
// valid hexadecimal digits (0-9, A-F, a-f).
func filterHexString(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if unicode.Is(unicode.ASCII_Hex_Digit, r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// hexToImage converts a hex string into an image using the ZX Spectrum palette.
// If width is zero, it will try to use a square image if possible, otherwise a single row.
func hexToImage(hexData string, width int) (image.Image, error) {
	// Remove optional "0x" or "0X" prefix.
	hexData = strings.TrimSpace(hexData)
	if strings.HasPrefix(hexData, "0x") || strings.HasPrefix(hexData, "0X") {
		hexData = hexData[2:]
	}

	// Filter out any non-hexadecimal characters (whitespace, tabs, cr, lf, etc).
	hexData = filterHexString(hexData)

	total := len(hexData)
	if total == 0 {
		return nil, errors.New("empty hex data")
	}

	// Determine width if not provided.
	if width == 0 {
		sq := int(math.Sqrt(float64(total)))
		if sq*sq == total {
			width = sq
		} else {
			width = total // single row.
		}
	}
	height := int(math.Ceil(float64(total) / float64(width)))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i, ch := range hexData {
		idx, err := strconv.ParseUint(string(ch), 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex digit '%c': %v", ch, err)
		}
		col := ZXPalette[idx]
		x := i % width
		y := i / width
		img.Set(x, y, col)
	}
	return img, nil
}

func saveImage(img image.Image, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}

// readHexFromFile reads the entire content of the file as a string.
func readHexFromFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// fileExists checks if a file exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	width := flag.Int("width", 0, "Width for output image when converting from hex data")
	output := flag.String("output", "", "Output filename")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: zxtex <input> [--width N] [--output file]")
		os.Exit(1)
	}

	input := flag.Arg(0)
	ext := strings.ToLower(filepath.Ext(input))
	var err error

	if fileExists(input) {
		switch ext {
		case ".png", ".jpg", ".jpeg", ".bmp", ".gif":
			hexStr, err := imageToHex(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting image: %v\n", err)
				os.Exit(1)
			}
			if *output != "" {
				f, err := os.Create(*output)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()
				writer := bufio.NewWriter(f)
				_, err = writer.WriteString(hexStr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
					os.Exit(1)
				}
				writer.Flush()
				fmt.Printf("Hex data written to %s\n", *output)
			} else {
				fmt.Println(hexStr)
			}
		case ".txt":
			hexData, err := readHexFromFile(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading hex file: %v\n", err)
				os.Exit(1)
			}
			img, err := hexToImage(hexData, *width)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting hex to image: %v\n", err)
				os.Exit(1)
			}
			outFile := *output
			if outFile == "" {
				outFile = "out.png"
			}
			err = saveImage(img, outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error saving image: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Image saved as %s\n", outFile)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported file type: %s\n", ext)
			os.Exit(1)
		}
	} else {
		// Treat input as a hex string.
		img, err := hexToImage(input, *width)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting hex string to image: %v\n", err)
			os.Exit(1)
		}
		outFile := *output
		if outFile == "" {
			outFile = "out.png"
		}
		err = saveImage(img, outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Image saved as %s\n", outFile)
	}
}
