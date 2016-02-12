// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
)

const xmin, ymin, xmax, ymax = -2.2, -1.2, +2.2, +1.2
const tile_size = 64

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	width, height, scaling, filepath := parse_cmdline()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	wait_group, worker_ch := spawn_processors(scaling, img)

	for py := 0; py < height; py += tile_size {
		for px := 0; px < width; px += tile_size {
			// (sur) should I worry about generating to many objects for GC ?
			worker_ch <- [2]int{py, px}
		}
	}
	// (sur) If we close a buffered channel with messages in the buffer,
	// do we lose them or are they going to be consumed by the other end ?
	close(worker_ch)
	wait_group.Wait()

	write_img(img, filepath)
}

func parse_cmdline() (width, height, scaling int, filepath string) {
	flag.IntVar(&width, "width", 2048, "Width of the resulting image")
	flag.IntVar(&height, "height", 1024, "Height of the resulting image")
	flag.IntVar(&scaling, "scaling", 4, "Oversampling factor of the image")
	flag.StringVar(&filepath, "filepath", "dupa.png",
		"File, where resulting image is going to be saved")

	flag.Parse()

	return
}

func write_img(img *image.RGBA, filepath string) {
	fo, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	w := bufio.NewWriter(fo)

	err = png.Encode(w, img) // NOTE: ignoring errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)
	}
}

func spawn_processors(subsample_f int, img *image.RGBA) (
	*sync.WaitGroup, chan<- [2]int) {
	// (sur) Should we worry about GC in general if we are sending too many
	// small elements through channel ?
	num_threads := runtime.NumCPU()
	worker_ch := make(chan [2]int, num_threads*2)
	var wg sync.WaitGroup

	for i := 0; i < num_threads; i++ {
		wg.Add(1)
		go tile_calculator(subsample_f, img, worker_ch, &wg, i)
	}

	return &wg, worker_ch
}

func tile_calculator(subsample_f int, img *image.RGBA, in <-chan [2]int,
	wg *sync.WaitGroup, thread_num int) {

	fmt.Fprintf(os.Stderr, "goroutine %d starting\n", thread_num)

	img_bounds := img.Bounds()
	y_delta := (ymax - ymin) / float64(img_bounds.Max.Y)
	x_delta := (xmax - xmin) / float64(img_bounds.Max.X)

	for d := range in {
		// (sur) What is the idiomatic way to unpack a list into vars ?
		// Something like:
		py, px := d[0], d[1]
		//fmt.Fprintf(os.Stderr, "Processing tile ((`%d`,`%d`),(`%d`,`%d`))\n",
		//py, px, py+tile_size, px+tile_size)
		var offset float64 = 0
		if subsample_f > 1 {
			offset = 0.5
		}

		y_base := (float64(py)-offset)*y_delta + ymin
		x_base := (float64(px)-offset)*x_delta + xmin

		for i_y := 0; i_y < tile_size; i_y++ {
			for i_x := 0; i_x < tile_size; i_x++ {
				var acumulator_blue, acumulator_red int64
				y := y_base + y_delta*float64(i_y)
				x := x_base + x_delta*float64(i_x)

				// Here's where the supersampling magic hapens:
				for ii_x := 0; ii_x < subsample_f; ii_x++ {
					for ii_y := 0; ii_y < subsample_f; ii_y++ {
						y += (y_delta / float64(subsample_f)) * float64(ii_y)
						x += (x_delta / float64(subsample_f)) * float64(ii_x)
						z := complex(x, y)
						blue, red := mandelbrot2(z)
						acumulator_blue += int64(blue)
						acumulator_red += int64(red)
					}
				}

				// (sur) How to idiomatically break this line ?
				avg_blue := uint8(float64(acumulator_blue) / float64(subsample_f*subsample_f))
				avg_red := uint8(float64(acumulator_red) / float64(subsample_f*subsample_f))
				img.Set(px+i_x, py+i_y, color.YCbCr{128, avg_blue, avg_red})
			}
		}
	}

	fmt.Fprintf(os.Stderr, "goroutine %d terminating\n", thread_num)
	wg.Done()
}

func mandelbrot2(z complex128) (blue, red uint8) {
	const iterations = 20000

	var v complex128
	for n := 0; n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			blue = uint8(real(v)*128) + 127
			red = uint8(imag(v)*128) + 127
			return blue, red
		}
	}
	return 0, 0
}

func acos(z complex128) color.Color {
	v := cmplx.Acos(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return color.YCbCr{192, blue, red}
}

func sqrt(z complex128) color.Color {
	v := cmplx.Sqrt(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return color.YCbCr{128, blue, red}
}

// f(x) = x^4 - 1
//
// z' = z - f(z)/f'(z)
//    = z - (z^4 - 1) / (4 * z^3)
//    = z - (z - 1/z^3) / 4
func newton(z complex128) color.Color {
	const iterations = 37
	const contrast = 7
	for i := uint8(0); i < iterations; i++ {
		z -= (z - 1/(z*z*z)) / 4
		if cmplx.Abs(z*z*z*z-1) < 1e-6 {
			return color.Gray{255 - contrast*i}
		}
	}
	return color.Black
}
