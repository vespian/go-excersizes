// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package algos

import "math/cmplx"

//(sur) Is it a problem that I make func type private and at the same time
// map using it public ?
type AlgoFunc func(r, i float64) (uint8, uint8)

// (sur) Is there a way to do _getattr_ in a simple way ? Or is the "map"
// approach idiomatic enough ?
var Str2funcMapping = map[string]AlgoFunc{
	"newton":         Newton,
	"acos":           Acos,
	"mandelbrotC64":  MandelbrotC64,
	"mandelbrotC128": MandelbrotC128,
	"sqrt":           Sqrt,
}

// Mandelbrot algo, but with complex128 type resolution
func MandelbrotC128(r, i float64) (uint8, uint8) {
	var z complex128 = complex(r, i)
	const iterations = 20000

	var v complex128
	for n := 0; n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			blue := uint8(real(v)*128) + 127
			red := uint8(imag(v)*128) + 127
			return blue, red
		}
	}
	return 0, 0
}

// Mandelbrot algo, but with complex64 type resolution
func MandelbrotC64(r, i float64) (uint8, uint8) {
	var z complex64 = complex(float32(r), float32(i))
	const iterations = 20000

	var v complex64
	for n := 0; n < iterations; n++ {
		v = v*v + z
		//(sur) I have not found a better way, golang seems to support only
		// float64 arithmetics :/
		if float32(cmplx.Abs(complex128(v))) > 2 {
			blue := uint8(real(v)*128) + 127
			red := uint8(imag(v)*128) + 127
			return blue, red
		}
	}
	return 0, 0
}

func Acos(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Acos(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

func Sqrt(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Sqrt(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

// f(x) = x^4 - 1
//
// z' = z - f(z)/f'(z)
//    = z - (z^4 - 1) / (4 * z^3)
//    = z - (z - 1/z^3) / 4
func Newton(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	const iterations = 20000

	for i := 0; i < iterations; i++ {
		z -= (z - 1/(z*z*z)) / 4
		if cmplx.Abs(z*z*z*z-1) < 1e-6 {
			blue := uint8(real(z)*128) + 127
			red := uint8(imag(z)*128) + 127
			return blue, red
		}
	}

	return 0, 0
}
