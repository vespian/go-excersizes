// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.

package img

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"sync"

	"github.com/vespian/go-excersizes/fractal_graphing/algos"
	"github.com/vespian/go-excersizes/fractal_graphing/common"
)

// (sur) One more question about breaking the line (here a func. signature) :)
func BuildImg(width, height, scaling int,
	algo func(r, i float64) (uint8, uint8)) *image.RGBA {

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	wait_group, worker_ch := spawnProcessors(scaling, img, algo)

	for py := 0; py < height; py += common.TILE_SIZE {
		for px := 0; px < width; px += common.TILE_SIZE {
			// (sur) should I worry about generating to many objects for GC ?
			worker_ch <- [2]int{py, px}
		}
	}

	close(worker_ch)
	wait_group.Wait()

	return img
}

func spawnProcessors(subsample_f int, img *image.RGBA, algo algos.AlgoFunc) (
	*sync.WaitGroup, chan<- [2]int) {
	// (sur) Should we worry about GC in general if we are sending too many
	// small elements through channel ?
	num_threads := runtime.NumCPU()
	worker_ch := make(chan [2]int, num_threads*2)
	var wg sync.WaitGroup

	for i := 0; i < num_threads; i++ {
		wg.Add(1)
		go tileCalculator(subsample_f, img, worker_ch, &wg, i, algo)
	}

	return &wg, worker_ch
}

//(sur) Is this the canonical way to break this line ?
func tileCalculator(subsample_f int,
	img *image.RGBA,
	in <-chan [2]int,
	wg *sync.WaitGroup, thread_num int,
	algo algos.AlgoFunc) {

	fmt.Fprintf(os.Stderr, "goroutine %d starting\n", thread_num)

	img_bounds := img.Bounds()
	y_delta := (common.YMAX - common.YMIN) / float64(img_bounds.Max.Y)
	x_delta := (common.XMAX - common.XMIN) / float64(img_bounds.Max.X)

	for d := range in {
		// (sur) What is the idiomatic way to unpack a list into vars ?
		// Something like: ?
		py, px := d[0], d[1]
		//fmt.Fprintf(os.Stderr, "Processing tile ((`%d`,`%d`),(`%d`,`%d`))\n",
		//py, px, py+common.TILE_SIZE, px+common.TILE_SIZE)
		var offset float64 = 0
		if subsample_f > 1 {
			offset = 0.5
		}

		y_base := (float64(py)-offset)*y_delta + common.YMIN
		x_base := (float64(px)-offset)*x_delta + common.XMIN

		for i_y := 0; i_y < common.TILE_SIZE; i_y++ {
			for i_x := 0; i_x < common.TILE_SIZE; i_x++ {
				var acumulator_blue, acumulator_red int64
				y := y_base + y_delta*float64(i_y)
				x := x_base + x_delta*float64(i_x)

				// Here's where the supersampling magic hapens:
				for ii_x := 0; ii_x < subsample_f; ii_x++ {
					for ii_y := 0; ii_y < subsample_f; ii_y++ {
						y += (y_delta / float64(subsample_f)) * float64(ii_y)
						x += (x_delta / float64(subsample_f)) * float64(ii_x)
						blue, red := algo(x, y)
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
