// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Package img contains all the code related to generating an image
package img

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"sync"

	"github.com/vespian/go-exercises/fractals/algos"
	"github.com/vespian/go-exercises/fractals/cmdline"
	"github.com/vespian/go-exercises/fractals/constants"
)

// BuildImg coordinates tasks between individual tile-processors. It does not
// directly write pixels, only sends tile numbers to calculate to processors.
func BuildImg(imgP *cmdline.ImgParams) *image.RGBA {

	f, _ := algos.MapStr2Func(imgP.Algo)
	img := image.NewRGBA(image.Rect(0, 0, imgP.Width, imgP.Height))
	waitGroup, workerCh := spawnProcessors(imgP.Scaling, img, f)

	for py := 0; py < imgP.Height; py += constants.TileSize {
		for px := 0; px < imgP.Width; px += constants.TileSize {
			// (sur) should I worry about generating to many objects for GC ?
			workerCh <- [2]int{py, px}
		}
	}

	close(workerCh)
	waitGroup.Wait()

	return img
}

func spawnProcessors(
	subsampleF int,
	img *image.RGBA,
	algo algos.AlgoFunc,
) (
	*sync.WaitGroup,
	chan<- [2]int,
) {

	numThreads := runtime.NumCPU()
	workerCh := make(chan [2]int, numThreads*2)
	var wg sync.WaitGroup

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go tileProcessor(subsampleF, img, workerCh, &wg, i, algo)
	}

	return &wg, workerCh
}

func tileProcessor(
	subsampleF int,
	img *image.RGBA,
	in <-chan [2]int,
	wg *sync.WaitGroup, threadNum int,
	algo algos.AlgoFunc,
) {

	fmt.Fprintf(os.Stderr, "goroutine %d starting\n", threadNum)

	imgBounds := img.Bounds()
	yDelta := (constants.YMax - constants.YMin) / float64(imgBounds.Max.Y)
	xDelta := (constants.XMax - constants.XMin) / float64(imgBounds.Max.X)

	for d := range in {
		//fmt.Fprintf(os.Stderr, "Processing tile ((`%d`,`%d`),(`%d`,`%d`))\n",
		//py, px, py+constants.TileSize, px+constants.TileSize)

		//py, px := d[0], d[1]
		calculateTile(yDelta, xDelta, d[0], d[1], subsampleF, img, algo)
	}

	fmt.Fprintf(os.Stderr, "goroutine %d terminating\n", threadNum)
	wg.Done()
}

func calculateTile(
	yDelta, xDelta float64,
	py, px int,
	subsampleF int,
	img *image.RGBA,
	algo algos.AlgoFunc,
) {
	var offset float64
	if subsampleF > 1 {
		offset = 0.5
	}
	yBase := (float64(py)-offset)*yDelta + constants.YMin
	xBase := (float64(px)-offset)*xDelta + constants.XMin

	for iY := 0; iY < constants.TileSize; iY++ {
		for iX := 0; iX < constants.TileSize; iX++ {
			var blue, red uint8
			y := yBase + yDelta*float64(iY)
			x := xBase + xDelta*float64(iX)
			if subsampleF > 1 {
				blue, red = calculateSupersampledPoint(subsampleF,
					y, x, yDelta, xDelta, algo)
			} else {
				blue, red = algo(x, y)
			}

			img.Set(px+iX, py+iY, color.YCbCr{128, blue, red})
		}
	}

}

func calculateSupersampledPoint(
	subsampleF int,
	y, x, yDelta, xDelta float64,
	algo algos.AlgoFunc,
) (
	uint8, uint8,
) {
	// Here's where the supersampling magic hapens:
	var acumulatorBlue, acumulatorRed int64

	for iiX := 0; iiX < subsampleF; iiX++ {
		for iiY := 0; iiY < subsampleF; iiY++ {
			y += (yDelta / float64(subsampleF)) * float64(iiY)
			x += (xDelta / float64(subsampleF)) * float64(iiX)
			blue, red := algo(x, y)
			acumulatorBlue += int64(blue)
			acumulatorRed += int64(red)
		}
	}

	avgBlue := uint8(float64(acumulatorBlue) / float64(subsampleF*subsampleF))
	avgRed := uint8(float64(acumulatorRed) / float64(subsampleF*subsampleF))

	return avgBlue, avgRed
}
