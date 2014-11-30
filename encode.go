package rgbe

import (
	"io"
	"bufio"
	"math"
	"fmt"
)

func Encode(w io.Writer, width, height int, data []float32) error {
	bw := bufio.NewWriter(w)

	if err := writeHeader(bw, width, height); err != nil {
		return err
	}

	if err := writePixels(bw, width * height, data); err != nil {
		return err
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

/*
int RGBE_WritePixels_RLE(FILE *fp, float *data, int scanline_width,
			 int num_scanlines)
{
  unsigned char rgbe[4];
  unsigned char *buffer;
  int i, err;

  if ((scanline_width < 8)||(scanline_width > 0x7fff))
    // run length encoding is not allowed so write flat
    return RGBE_WritePixels(fp,data,scanline_width*num_scanlines);
  buffer = (unsigned char *)malloc(sizeof(unsigned char)*4*scanline_width);
  if (buffer == NULL) 
    // no buffer space so write flat 
    return RGBE_WritePixels(fp,data,scanline_width*num_scanlines);
  while(num_scanlines-- > 0) {
    rgbe[0] = 2;
    rgbe[1] = 2;
    rgbe[2] = scanline_width >> 8;
    rgbe[3] = scanline_width & 0xFF;
    if (fwrite(rgbe, sizeof(rgbe), 1, fp) < 1) {
      free(buffer);
      return rgbe_error(rgbe_write_error,NULL);
    }
    for(i=0;i<scanline_width;i++) {
      float2rgbe(rgbe,data[RGBE_DATA_RED],
		 data[RGBE_DATA_GREEN],data[RGBE_DATA_BLUE]);
      buffer[i] = rgbe[0];
      buffer[i+scanline_width] = rgbe[1];
      buffer[i+2*scanline_width] = rgbe[2];
      buffer[i+3*scanline_width] = rgbe[3];
      data += RGBE_DATA_SIZE;
    }
    // write out each of the four channels separately run length encoded 
    // first red, then green, then blue, then exponent 
    for(i=0;i<4;i++) {
      if ((err = RGBE_WriteBytes_RLE(fp,&buffer[i*scanline_width],
				     scanline_width)) != RGBE_RETURN_SUCCESS) {
	free(buffer);
	return err;
      }
    }
  }
  free(buffer);
  return RGBE_RETURN_SUCCESS;
}
*/

func writePixels_RLE(w *bufio.Writer, scanlineWidth, numScanlines int, data []float32) error {
	if scanlineWidth < 8 || scanlineWidth > 0x7fff {
		// run length encoding is not allowed so write flat
		return writePixels(w, scanlineWidth * numScanlines, data)
	}

	for ; numScanlines > 0; numScanlines-- {

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