rgbe
====

Description
-----------

rgbe is a simple RGBE encoding/decoding library for Go. The RGBE format was invented by Greg Ward for compact storage of HDR images (http://radsite.lbl.gov/radiance/refer/Notes/picture_format.html). This library is based on code written by Bruce Walter (http://www.graphics.cornell.edu/~bjw/rgbe/).

Reading of both the RLE-compressed and uncompressed variants is supported. Images are always written with RLE-comression. No additional header information (e.g. gamma and exposure) are read or written.

Pixel data is stored in RGB order in a []float32 slice, starting in the left top corner and going downards from left to right.

Example
-------

This example reads a RGBE encoded image from disk, computes the average luminance and stores a black and white variant of the image in RGBE format on  disk again. Error handling is ommitted for brevity.

```Go
package main

import (
	"github.com/Opioid/rgbe"
	"os"
	"fmt"
)

func main() {
	fi, _ := os.Open("output.hdr")
	defer fi.Close()

	width, height, data, _ := rgbe.Decode(fi)

	total := float32(width * height)
	averageLuminance := float32(0)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			o := (width * y + x) * 3
			r, g, b := data[o], data[o + 1], data[o + 2]
			l := luminance(r, g, b)
			averageLuminance += l / total
			data[o], data[o + 1], data[o + 2] = l, l, l
		}
	}

	fmt.Printf(
		"RGBE image %d x %d size, average luminance is %v\n", 
		width, height, averageLuminance)

	fo, _ := os.Create("bw.hdr")
	defer fo.Close()

	rgbe.Encode(fo, width, height, data)
}

func luminance(r, g, b float32) float32 {
		return r * 0.299 + g * 0.587 + b * 0.114
}
```

Tasks
-----

- [ ] Support additional header information (e.g. Gamma, Exposure)?
- [ ] Write tests