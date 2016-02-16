// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package common

import (
	"fmt"

	"github.com/vespian/go-excersizes/fractal_graphing/algos"
)

const XMIN, YMIN, XMAX, YMAX = -2.2, -1.1, +2.2, +1.1
const TILE_SIZE = 64
const DEFAULT_WIDTH, DEFAULT_HEIGHT = 2048, 1024
const DEFAULT_SCALING = 1
const DEFAULT_ALGO = "mandelbrotC128"

func ValidateImgParams(width, height, scaling int, algo string) error {
	if width%TILE_SIZE != 0 || height%TILE_SIZE != 0 {
		msg_fmt := "Width(%d) and height(%d) of the resulting picture" +
			" should be multiples of tile size(%d)\n"
		return fmt.Errorf(msg_fmt, width, height, TILE_SIZE)
	}

	ratio_xy := float64(YMAX-YMIN) / float64(XMAX-XMIN)
	ratio_pxpy := float64(height) / float64(width)
	if ratio_xy != ratio_pxpy {
		sugested_width := int(float64(height) / ratio_xy)
		msg_fmt := "Pixel ratio (%2.2f) differs from XY ratio(%2.2f), try" +
			" adjusting width to %d\n\n"
		return fmt.Errorf(msg_fmt, ratio_pxpy, ratio_xy, sugested_width)
	}

	if scaling < 1 {
		msg_fmt := "Scaling factor must be >= 1, currently: `%d`\n"
		return fmt.Errorf(msg_fmt, scaling)
	}

	if _, ok := algos.Str2funcMapping[algo]; !ok {
		msg := "algorithm must be one of: mandelbrot_(c64|c128)|sqrt|acos" +
			"|newton, given: %s\n"
		return fmt.Errorf(msg, algo)
	}

	return nil

}
