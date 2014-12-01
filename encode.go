package rgbe

// http://www.graphics.cornell.edu/~bjw/rgbe

import (
	"io"
	"bufio"
	"math"
	"fmt"
)

const (
	minRunLength = 4
)

func Encode(w io.Writer, width, height int, data []float32) error {
	bw := bufio.NewWriter(w)

	if err := writeHeader(bw, width, height); err != nil {
		return err
	}

	useRLE := true

	if useRLE {
		if err := writePixels_RLE(bw, width, height, data); err != nil {
			return err
		}
	} else {
		if err := writePixels(bw, width * height, data); err != nil {
			return err
		}
	}

	if err := bw.Flush(); err != nil {
		return newError(WriteError, err.Error())
	}

	return nil
}

func writeHeader(w *bufio.Writer, width, height int) error {
	if _, err := w.WriteString("#?RGBE\n"); err != nil {
		return newError(WriteError, err.Error())
	}

	if _, err := w.WriteString("FORMAT=32-bit_rle_rgbe\n\n"); err != nil {
		return newError(WriteError, err.Error())
	}

	line := fmt.Sprintf("-Y %d +X %d\n", height, width)

	if _, err := w.WriteString(line); err != nil {
		return newError(WriteError, err.Error())
	}

	return nil
}


// simple write routine that does not use run length encoding 
func writePixels(w *bufio.Writer, numPixels int, data []float32) error {
	rgbe := make([]byte, 4)

	index := 0

	for ; numPixels > 0; numPixels-- {
		r := data[index]
		g := data[index + 1]
		b := data[index + 2]

		floatToRgbe(r, g, b, rgbe)

		index += 3

		if _, err := w.Write(rgbe); err != nil {
			return newError(WriteError, err.Error())
		}
	}

	return nil
}

func writePixels_RLE(w *bufio.Writer, scanlineWidth, numScanlines int, data []float32) error {
	if scanlineWidth < 8 || scanlineWidth > 0x7fff {
		// run length encoding is not allowed so write flat
		return writePixels(w, scanlineWidth * numScanlines, data)
	}

	index := 0
	rgbe := make([]byte, 4)
	scanlineBuffer := make([]byte, 4 * scanlineWidth)

	for ; numScanlines > 0; numScanlines-- {
		rgbe[0] = 2
		rgbe[1] = 2
		rgbe[2] = byte(scanlineWidth >> 8)
		rgbe[3] = byte(scanlineWidth & 0xFF)

		if _, err := w.Write(rgbe); err != nil {
			return newError(WriteError, err.Error())
		}

		for i := 0; i < scanlineWidth; i++ {
			r := data[index]
			g := data[index + 1]
			b := data[index + 2]

			floatToRgbe(r, g, b, rgbe)

			index += 3

			scanlineBuffer[i]                     = rgbe[0]
			scanlineBuffer[i + scanlineWidth]     = rgbe[1]
			scanlineBuffer[i + 2 * scanlineWidth] = rgbe[2]
			scanlineBuffer[i + 3 * scanlineWidth] = rgbe[3]
		}

		for i := 0; i < 4; i++ {
			start := i * scanlineWidth
			if err := writeBytes_RLE(w, scanlineBuffer[start:start + scanlineWidth]); err != nil {
				return newError(WriteError, err.Error())
			}
		}
	}

	return nil
}

// The code below is only needed for the run-length encoded files.
// Run length encoding adds considerable complexity but does
// save some space.  For each scanline, each channel (r,g,b,e) is
// encoded separately for better compression.
func writeBytes_RLE(w *bufio.Writer, data []byte) error {
	numBytes := len(data)
	buf := make([]byte, 2)
	cur := 0

	for cur < numBytes {
		begRun := cur

		// find next run of length at least 4 if one exists
		runCount, oldRunCount := 0, 0

		for runCount < minRunLength && begRun < numBytes {
			begRun += runCount
			oldRunCount = runCount
			runCount = 1

			for ; begRun + runCount < numBytes && runCount < 127 && data[begRun] == data[begRun + runCount]; runCount++ {
			}
		}

		// if data before next big run is a short run then write it as such
		if oldRunCount > 1 && oldRunCount == begRun - cur {
			buf[0] = byte(128 + oldRunCount) // write short run
			buf[1] = data[cur]

			if _, err := w.Write(buf); err != nil {
				return newError(WriteError, err.Error())
			}

			cur = begRun
		}

		// write out bytes until we reach the start of the next run
		for cur < begRun {
			nonRunCount := begRun - cur

			if nonRunCount > 128 {
				nonRunCount = 128
			}

			if err := w.WriteByte(byte(nonRunCount)); err != nil {
				return newError(WriteError, err.Error())
			}

			if _, err := w.Write(data[cur:cur + nonRunCount]); err != nil {
				return newError(WriteError, err.Error())
			}

			cur += nonRunCount
		}

		// write out next run if one was found
		if runCount >= minRunLength {
			buf[0] = byte(128 + runCount)
			buf[1] = data[begRun]

			if _, err := w.Write(buf); err != nil {
				return newError(WriteError, err.Error())
			}

			cur += runCount
		} 
	}

	return nil
}

// standard conversion from float pixels to rgbe pixels 
func floatToRgbe(r, g, b float32, rgbe []byte) {
	v := r

	if g > v {
		v = g
	}

	if b > v {
		v = b
	}

	if v < 1e-32 {
		rgbe[0], rgbe[1], rgbe[2], rgbe[3] = 0, 0, 0, 0
	} else {
		f, e := math.Frexp(float64(v))
		f32 := float32(f) * 256 / v

		rgbe[0] = byte(r * f32)
		rgbe[1] = byte(g * f32)
		rgbe[2] = byte(b * f32)
		rgbe[3] = byte(e + 128)
	}
}