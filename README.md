# zxtex

**zxtex** is a command-line tool for converting images (PNG, GIF, and BMP) to and from a simple, text-based hex sprite format inspired by the ZX Spectrum palette. In this format, each pixel is represented by a single hexadecimal digit (0–F) or by a dot (`.`) for transparent pixels.

## Features

- **Image-to-Hex Conversion:**  
  Convert PNG, GIF, or BMP images to a text file containing hex data.
  - **Row Mode (default):**  
    Outputs header metadata and one line per image row. The header includes the original filename, width, height, and generator info.
  - **Raw Mode:**  
    Use the `--raw` flag to output a single continuous hex string with no header or newlines (a newline is appended at the end).

- **Hex-to-Image Conversion:**  
  Convert a hex text file (with a `.txt` or `.hex` extension) or a direct hex string back into a PNG image.
  - When reading a text file, header lines (starting with `#`) are ignored, and the width is taken from the first non-empty line if not specified.
  - If the header contains an original filename (from a line like `# file: invader.png`) and no output filename is specified, the output image will be saved using that base name with a `.png` extension.

- **Direct String Mode:**  
  You can also pass a continuous hex string directly as an argument. In this mode, the `--width` flag is mandatory.

- **Transparency Support and Overrides:**  
  Fully transparent pixels are represented by the dot character (`.`) in the hex format.  
  Additionally, you can override transparency for images that do not support an alpha channel by using one of the following flags:
  - `--transpcolor` or `--transpcolour`: Provide a web-format color (e.g. `#aabbcc`) that should be treated as transparent.
  - `--transpindex`: Specify a palette index (an integer) to treat as transparent.

## Installation

Make sure you have [Go](https://golang.org) installed. Then, clone the repository and build the binary:

```bash
git clone https://github.com/ha1tch/zxtex.git
cd zxtex
go get golang.org/x/image/bmp
go build zxtex.go
```

This creates the `zxtex` executable.

## Usage

```
Usage: zxtex <input> [--raw] [--width N] [--output file] [--transpcolor #aabbcc|--transpindex N]
```

- `<input>`: Can be an image file (PNG, GIF, BMP), a text file (`.txt` or `.hex`), or a direct hex string.
- `--raw`: (Optional) When converting an image to hex, outputs a single continuous hex string (no header, no newlines). A newline is appended at the end.
- `--width N`: (Mandatory in direct string mode or if the hex file does not specify a width) Specifies the width (in pixels) for image reconstruction.
- `--output file`: (Optional) Specifies the output filename. For hex-to-image conversion, if no output filename is provided and the hex file header contains an original filename, the output file will use that name (with a `.png` extension).
- `--transpcolor` / `--transpcolour`: (Optional) Specifies a web color (e.g. `#aabbcc`) that should be interpreted as transparent.
- `--transpindex`: (Optional) Specifies a palette index that should be interpreted as transparent.

### Examples

#### Convert an Image to Hex (Row Mode)

```bash
./zxtex invader.png
```

_Output (to standard output):_

```
# file: invader.png
# width: 13
# height: 8
# generator: zxtex
...7.....7...
....7...7....
...7777777...
..770777077..
.77777777777.
.7.7777777.7.
.7.7.....7.7.
....7...7....
```

#### Convert an Image to Hex (Raw Mode)

```bash
./zxtex --raw invader.png
```

_Output (a single continuous string, ending with a newline):_

```
...7.....7....7...7....7777777...770777077..77777777777..7.7777777.7..7.7.....7.7....7...7....
```

#### Convert a Hex File to an Image

If you have a hex file `invader.hex` (with header metadata) and want to generate a PNG, run:

```bash
./zxtex --output myinvader.png invader.hex
```

If the hex file header contains an original filename and no output filename is specified, zxtex will use the base name from the header (with a `.png` extension).

#### Direct Hex String to Image

```bash
./zxtex --width 13 "....7...7....7777777..."
```

Make sure to specify the width when using a direct hex string as input.

#### Override Transparency

For input images that do not support transparency, you can force a specific color or palette index to be treated as transparent. For example:

- Using a web-format color:

  ```bash
  ./zxtex --transpcolor "#ff00ff" --output out.hex image.bmp
  ```

- Using a palette index:

  ```bash
  ./zxtex --transpindex 2 --raw image.bmp
  ```

## License

This project is licensed under the Apache License 2.0.

## Author

**haitchfive**  
Email: haitch@duck.com  
Social media: [https://oldbytes.space/@haitchfive](https://oldbytes.space/@haitchfive)

## Repository

GitHub: [https://github.com/ha1tch/zxtex.git](https://github.com/ha1tch/zxtex.git)


