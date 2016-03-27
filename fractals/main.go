// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/vespian/go-exercises/fractals/cmdline"
	"github.com/vespian/go-exercises/fractals/img"
	"github.com/vespian/go-exercises/fractals/web"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cmdargs := cmdline.Cmdline()
	if cmdargs.Err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse commandline: %s", cmdargs.Err)
		os.Exit(1)
	}

	if cmdargs.Filepath != "" {
		// algo is already validated by cmdline.Cmdline
		img := img.BuildImg(&cmdargs.ImgParams)
		if err := cmdline.WriteImg(img, cmdargs.Filepath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write Image to file: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if cmdargs.Ssocket != "" {
		http.HandleFunc("/", web.ServeHTTP)
		// Forever
		err := http.ListenAndServe(cmdargs.Ssocket, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start http server: %s", err)
			os.Exit(1)
		}
	}

	// Never reached, cmdline.Cmdline() makes sure that either cmdargs.Filepath
	// or cmdargs.Ssocket evaluates to true.
}
