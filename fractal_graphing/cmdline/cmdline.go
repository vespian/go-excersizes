// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package cmdline

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/vespian/go-excersizes/fractal_graphing/common"
)

func parseCmdline() (width, height, scaling int, filepath, ssocket, algo string) {
	flag.IntVar(&width, "width", common.DEFAULT_WIDTH, "Width of the resulting image")
	flag.IntVar(&height, "height", common.DEFAULT_HEIGHT, "Height of the resulting image")
	flag.IntVar(&scaling, "scaling", common.DEFAULT_SCALING,
		"Super-sampling factor of the image (1 == no super-sampling)")
	flag.StringVar(&filepath, "filepath", "",
		"File, where resulting image is going to be saved")
	flag.StringVar(&ssocket, "ssocket", "",
		"Http server socket")
	//(sur) Is this the idiomatic way how to break/format strings ?
	flag.StringVar(&algo, "algorithm", common.DEFAULT_ALGO,
		"Algorithm to use to calculate the fractal (newton|acos|mandelbrot"+
			"(_c64|_c128)|sqrt)")

	flag.Parse()

	return
}

func validateCmdline(width, height, scaling int, filepath, ssocket, algo string) {
	if filepath == "" && ssocket == "" {
		msg := "Either path to an output file or server socket " +
			"must be specified\n"
		fmt.Fprintf(os.Stderr, msg)
		os.Exit(1)
	}

	if filepath != "" && ssocket != "" {
		msg := "Filepath and ssocket are mutually exclusive\n"
		fmt.Fprintf(os.Stderr, msg)
		os.Exit(1)
	}

	if filepath != "" {
		err := common.ValidateImgParams(width, height, scaling, algo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
	} else {
		// Validated while handling HTTP request
	}

}

func GetCmdline() (width, height, scaling int, filepath, ssocket, algo string) {
	width, height, scaling, filepath, ssocket, algo = parseCmdline()
	validateCmdline(width, height, scaling, filepath, ssocket, algo)
	return
}
func WriteImg(img *image.RGBA, filepath string) {
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
