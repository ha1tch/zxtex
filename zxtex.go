package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "golang.org/x/image/bmp" // register BMP format
	_ "image/gif"              // register GIF format
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

// Global flags for transparency override.
var transpColorStr string
var transpIndex int

// parseWebColor parses a web-format color string (e.g. "#aabbcc") and returns a color.RGBA.
func parseWebColor(s string) (color.RGBA, error) {
	// Remove leading '#' if present.
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid web color %q: must be 6 hex digits", s)
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}

// nearestColor returns the index of the nearest ZX Spectrum palette color for the given color.
func nearestColor(r, g, b uint32) int {
	bestIndex := 0
	bestDist := math.MaxFloat64
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

// shouldBeTransparent returns true if the pixel should be treated as transparent.
// It checks if alpha is 0 or if it matches the user-specified transparent color or palette index.
func shouldBeTransparent(r, g, b, a uint32) bool {
	// a is 16-bit; fully opaque is 0xFFFF.
	if a == 0 {
		return true
	}

	// If a transparent color is specified, compare 8-bit values.
	if transpColorStr != "" {
		tcol, err := parseWebColor(transpColorStr)
		if err == nil {
			// Convert pixel to 8-bit.
			pr := uint8(r >> 8)
			pg := uint8(g >> 8)
			pb := uint8(b >> 8)
			if pr == tcol.R && pg == tcol.G && pb == tcol.B {
				return true
			}
		}
	}

	// If a transparent palette index is specified (>=0), use nearestColor.
	if transpIndex >= 0 {
		idx := nearestColor(r, g, b)
		if idx == transpIndex {
			return true
		}
	}

	return false
}

// imageToHex converts an image file into a hex string with header metadata and one line per row.
func imageToHex(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	if format != "png" && format != "gif" && format != "bmp" {
		return "", fmt.Errorf("unsupported image format: %s (only PNG, GIF, and BMP are supported)", format)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	var sb strings.Builder
	// Header metadata.
	sb.WriteString(fmt.Sprintf("# file: %s\n", filename))
	sb.WriteString(fmt.Sprintf("# width: %d\n", width))
	sb.WriteString(fmt.Sprintf("# height: %d\n", height))
	sb.WriteString("# generator: zxtex\n")
	// One line per row.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var rowBuilder strings.Builder
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			if shouldBeTransparent(r, g, b, a) {
				rowBuilder.WriteRune('.')
			} else {
				idx := nearestColor(r, g, b)
				rowBuilder.WriteString(strings.ToUpper(strconv.FormatInt(int64(idx), 16)))
			}
		}
		sb.WriteString(rowBuilder.String())
		sb.WriteRune('\n')
	}
	return sb.String(), nil
}

// imageToRawHex converts an image file into a single continuous hex string (no header, no newlines).
func imageToRawHex(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	if format != "png" && format != "gif" && format != "bmp" {
		return "", fmt.Errorf("unsupported image format: %s (only PNG, GIF, and BMP are supported)", format)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	var sb strings.Builder
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			if shouldBeTransparent(r, g, b, a) {
				sb.WriteRune('.')
			} else {
				idx := nearestColor(r, g, b)
				sb.WriteString(strings.ToUpper(strconv.FormatInt(int64(idx), 16)))
			}
		}
	}
	sb.WriteRune('\n') // Append a newline at the end.
	return sb.String(), nil
}

// filterHexLine removes spaces and tabs from a line, but keeps the dot.
func filterHexLine(line string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, line)
}

// filterHexString removes all characters that are not valid hex digits or the '.' placeholder.
func filterHexString(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if unicode.Is(unicode.ASCII_Hex_Digit, r) || r == '.' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// readHexFromTextFile reads a text file (which may include header comments) and returns a continuous hex string,
// the width (from the first non-empty line), and the original filename from the header (if any).
func readHexFromTextFile(filename string) (string, int, string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", 0, "", err
	}
	content := string(bytes)
	scanner := bufio.NewScanner(strings.NewReader(content))
	var filteredLines []string
	width := 0
	origFileName := ""
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")
		// Check for header lines.
		if strings.HasPrefix(line, "#") {
			// Look for the original filename in a header like "# file: invader.png"
			if strings.HasPrefix(strings.ToLower(line), "# file:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					origFileName = strings.TrimSpace(parts[1])
				}
			}
			continue
		}
		// Remove inline comments.
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		filtered := filterHexLine(line)
		if len(filtered) > 0 {
			if width == 0 {
				width = len(filtered)
			}
			filteredLines = append(filteredLines, filtered)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", 0, "", err
	}
	joined := strings.Join(filteredLines, "")
	joined = filterHexString(joined)
	return joined, width, origFileName, nil
}

// hexToImage converts a continuous hex string into an image.
func hexToImage(hexData string, width int) (image.Image, error) {
	total := len(hexData)
	if total == 0 {
		return nil, errors.New("empty hex data")
	}
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
		x := i % width
		y := i / width
		if ch == '.' {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		} else {
			idx, err := strconv.ParseUint(string(ch), 16, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid hex digit '%c': %v", ch, err)
			}
			col := ZXPalette[idx]
			img.Set(x, y, col)
		}
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	rawMode := flag.Bool("raw", false, "Output as a single continuous hex string with no header or row breaks")
	widthFlag := flag.Int("width", 0, "Width for output image when converting from hex data (mandatory in direct string mode)")
	output := flag.String("output", "", "Output filename")
	// New flags for transparent colour override.
	transpColorFlag := flag.String("transpcolor", "", "Transparent color (in web format, e.g. #aabbcc) to use as transparent")
	transpColourFlag := flag.String("transpcolour", "", "Transparent colour (in web format, e.g. #aabbcc) to use as transparent")
	transpIndexFlag := flag.Int("transpindex", -1, "Palette index to treat as transparent")
	flag.Parse()

	// Use either transpcolor or transpcolour if provided.
	if *transpColorFlag != "" {
		transpColorStr = *transpColorFlag
	} else if *transpColourFlag != "" {
		transpColorStr = *transpColourFlag
	}
	transpIndex = *transpIndexFlag

	if flag.NArg() < 1 {
		fmt.Println("Usage: zxtex <input> [--raw] [--width N] [--output file] [--transpcolor #aabbcc|--transpindex N]")
		os.Exit(1)
	}

	input := flag.Arg(0)
	ext := strings.ToLower(filepath.Ext(input))
	if fileExists(input) {
		switch ext {
		// If input is an image, convert it to hex.
		case ".png", ".gif", ".bmp":
			var hexStr string
			var err error
			if *rawMode {
				hexStr, err = imageToRawHex(input)
			} else {
				hexStr, err = imageToHex(input)
			}
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
				fmt.Print(hexStr)
			}
		// If input is a text file, read it and convert to an image.
		case ".txt", ".hex":
			hexData, fileWidth, origName, err := readHexFromTextFile(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading hex file: %v\n", err)
				os.Exit(1)
			}
			useWidth := *widthFlag
			if useWidth == 0 && fileWidth > 0 {
				useWidth = fileWidth
			}
			img, err := hexToImage(hexData, useWidth)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting hex to image: %v\n", err)
				os.Exit(1)
			}
			outFile := *output
			if outFile == "" {
				// If an original filename is available in metadata, use its base name with a .png extension.
				if origName != "" {
					base := filepath.Base(origName)
					ext := filepath.Ext(base)
					nameOnly := strings.TrimSuffix(base, ext)
					outFile = nameOnly + ".png"
				} else {
					outFile = "out.png"
				}
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
		// Direct string mode.
		if *widthFlag == 0 {
			fmt.Fprintln(os.Stderr, "In direct string mode, you must specify the --width flag.")
			os.Exit(1)
		}
		hexStr := strings.TrimSpace(input)
		if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
			hexStr = hexStr[2:]
		}
		hexStr = filterHexString(hexStr)
		img, err := hexToImage(hexStr, *widthFlag)
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
