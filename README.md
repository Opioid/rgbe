rgbe
====

Description
-----------

rgbe is a RGBE decoder for Go. The RGBE format was invented by Greg Ward for compact storage of HDR images. This library is based on code by Bruce Walter (http://www.graphics.cornell.edu/~bjw/rgbe/).

At the moment the only supported use case is reading from an io.Reader interface and storing the result in []float32. 

Example
-------

This example opens a RGBE encoded file from disk, computes the average color of the image and prints the result.

```Go
package main

import (
	"github.com/Opioid/rgbe"
	"os"
	"fmt"
)

func main() {
	fi, err := os.Open("example.hdr")

	defer fi.Close()

	if err != nil {
		panic(err)
	}

	width, height, data, err := rgbe.Decode(fi)

	if err != nil {
		panic(err)
	}

	total := float32(width * height)

	r, g, b := float32(0), float32(0), float32(0)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			o := (width * y + x) * 3

			r += data[o + 0] / total
			g += data[o + 1] / total
			b += data[o + 2] / total
		}
	}

	fmt.Printf("RGBE image %d x %d size, average color [%v, %v, %v]\n", width, height, r, g, b)
}
```

Tasks
-----

- [ ] Support encoding
- [ ] Support additional header information (e.g. Gamma, Exposure)
- [ ] Write tests