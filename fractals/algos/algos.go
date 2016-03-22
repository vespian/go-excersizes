// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Package algos groups functions used for calculating fractals
package algos

import (
	"fmt"
	"math/cmplx"
)

// AlgoFunc is the signature of the fuctions used for generating fractals
type AlgoFunc func(r, i float64) (uint8, uint8)

var str2funcMapping = map[string]AlgoFunc{
	"newton":         Newton,
	"acos":           Acos,
	"mandelbrotC64":  MandelbrotC64,
	"mandelbrotC128": MandelbrotC128,
	"sqrt":           Sqrt,
}

// MapStr2Func converts string name of the fractal generator into a reference
// of the function implementing it.
func MapStr2Func(algo string) (AlgoFunc, error) {
	var val AlgoFunc

	if _, ok := str2funcMapping[algo]; !ok {
		msg := "algorithm must be one of: \n"
		for key := range str2funcMapping {
			msg += fmt.Sprintf(" - %s\n", key)
		}
		msg += fmt.Sprintf("Given: %s\n", algo)
		return nil, fmt.Errorf(msg)
	}
	return val, nil
}

// MandelbrotC128 calculates pixel values for Mandelbrot fractal using
// complex128 type.
func MandelbrotC128(r, i float64) (uint8, uint8) {
	const iterations = 20000

	var v complex128
	z := complex(r, i)

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

// MandelbrotC64 calculates pixel values for Mandelbrot fractal using
// complex64 type(at least tries to :) ).
func MandelbrotC64(r, i float64) (uint8, uint8) {
	const iterations = 20000

	var v complex64
	z := complex(float32(r), float32(i))

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

// Acos calculates pixel values for arcus-cosinus fractal.
func Acos(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Acos(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

// Sqrt calculates pixel values for sqrt fractal.
func Sqrt(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Sqrt(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

// Newton calculates pixel values for Newton's method of finding minimas.
//
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
