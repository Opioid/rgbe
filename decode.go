package rgbe

// http://www.graphics.cornell.edu/~bjw/rgbe

import (
	"io"
	"bufio"
	"math"
	"fmt"
)

func Decode(r io.Reader) (int, int, []float32, error) {
	br := bufio.NewReader(r)

	width, height, err := readHeader(br) 

	if err != nil {
		return 0, 0, nil, err
	} 

	data := make([]float32, width * height * 3)

	if err := readPixels_RLE(br, width, height, data); err != nil {
		return 0, 0, nil, err
	}

	return width, height, data, nil
}

func readHeader(r *bufio.Reader) (int, int, error) {
	line, err := r.ReadString('\n')

	if err != nil {
		return 0, 0, newError(ReadError, err.Error())
	}

	if line[0] != '#' || line[1] != '?' {
		return 0, 0, newError(FormatError, "Bad initial token.")
	}

	formatSpecifier := false

	for {
		line, err = r.ReadString('\n')

		if err != nil {
			return 0, 0, newError(ReadError, err.Error())
		}

		if line[0] == 0 || line[0] == '\n' {
			// blank lines signifies end of meta data header
			break
		} else if line == "FORMAT=32-bit_rle_rgbe\n" {
			formatSpecifier = true
		}
	}	

	if !formatSpecifier {
		return 0, 0, newError(FormatError, "No FORMAT specifier found.")
	}

	line, err = r.ReadString('\n')

	if err != nil {
		return 0, 0, newError(ReadError, err.Error())
	}

	width, height := 0, 0
	if n, err := fmt.Sscanf(line, "-Y %d +X %d", &height, &width); n < 2 || err != nil {
		return 0, 0, newError(FormatError, "Missing image size specifier.")
	}

	return width, height, nil
}

func readPixels_RLE(r *bufio.Reader, scanlineWidth, numScanlines int, data []float32) error {
	if scanlineWidth < 8 || scanlineWidth > 0x7fff {
		// run length encoding is not allowed so read flat
		return readPixels(r, scanlineWidth * numScanlines, data)
	} 

	offset := 0
	rgbe := make([]byte, 4)
	scanlineBuffer := make([]byte, 4 * scanlineWidth)
	buf := make([]byte, 2)

	for ; numScanlines > 0; numScanlines-- {
		if _, err := io.ReadFull(r, rgbe); err != nil {
			return newError(ReadError, err.Error())
		}

		if rgbe[0] != 2 || rgbe[1] != 2 || (rgbe[2] & 0x80) != 0 {
			// this file is not run length encoded
			data[0], data[1], data[2] = rgbeToFloat(rgbe[0], rgbe[1], rgbe[2], rgbe[3])

			return readPixels(r, scanlineWidth * numScanlines - 1, data[3:])
		}

		if int(rgbe[2]) << 8 | int(rgbe[3]) != scanlineWidth {
			return newError(FormatError, "Wrong scanline width.")
		}

		// read each of the four channels for the scanline into the buffer 
		index := 0
		for i := 0; i < 4; i++ {
			end := (i + 1) * scanlineWidth

			for index < end {
				if _, err := io.ReadFull(r, buf); err != nil {
					return newError(ReadError, err.Error())
				}

				if buf[0] > 128 {
					 // a run of the same value
					count := int(buf[0]) - 128

					if count == 0 || count > end - index {
						return newError(FormatError, "Bad scanline data.")
					}

					for ; count > 0; count-- {
						scanlineBuffer[index] = buf[1]
						index++
					}
				} else {
					// a non-run
					count := int(buf[0])

					if count == 0 || count > end - index {
						return newError(FormatError, "Bad scanline data.")
					}

					scanlineBuffer[index] = buf[1]
					index++

					count--
					if count > 0 {
						if _, err := io.ReadFull(r, scanlineBuffer[index:index + count]); err != nil {
							return newError(ReadError, err.Error())
						}

						index += count
					} 
				}
			}
		}

		// now convert data from buffer into floats 
		for i := 0; i < scanlineWidth; i++ {
			r := scanlineBuffer[i]
			g := scanlineBuffer[i +     scanlineWidth]
			b := scanlineBuffer[i + 2 * scanlineWidth]
			e := scanlineBuffer[i + 3 * scanlineWidth]

			data[offset], data[offset + 1], data[offset + 2] = rgbeToFloat(r, g, b, e)
			offset += 3
		}
	}

	return nil
}

func readPixels(r *bufio.Reader, numPixels int, data []float32) error {
	rgbe := make([]byte, 4)
	offset := 0

	for ; numPixels > 0; numPixels-- {
		if _, err := io.ReadFull(r, rgbe); err != nil {
			return newError(MemoryError, err.Error())
		}

		data[offset], data[offset + 1], data[offset + 2] = rgbeToFloat(rgbe[0], rgbe[1], rgbe[2], rgbe[3])
		offset += 3
	}

	return nil
}

// standard conversion from rgbe to float pixels 
// note: Ward uses ldexp(col+0.5,exp-(128+8)).  However we wanted pixels 
//       in the range [0,1] to map back into the range [0,1].            
func rgbeToFloat(r, g, b, e byte) (float32, float32, float32) {
	if e > 0 {
		// nonzero pixel
		f := float32(math.Ldexp(1, int(e) - (128 + 8)))

		return float32(r) * f, float32(g) * f, float32(b) * f
	} else {
		return 0, 0, 0
	}
}