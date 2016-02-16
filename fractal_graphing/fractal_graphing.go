// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/vespian/go-excersizes/fractal_graphing/algos"
	"github.com/vespian/go-excersizes/fractal_graphing/cmdline"
	fghttp "github.com/vespian/go-excersizes/fractal_graphing/http"
	"github.com/vespian/go-excersizes/fractal_graphing/img"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	width, height, scaling, filepath, ssocket, algo := cmdline.GetCmdline()

	if filepath != "" {
		f := algos.Str2funcMapping[algo]
		img := img.BuildImg(width, height, scaling, f)
		cmdline.WriteImg(img, filepath)
	}

	if ssocket != "" {
		http.HandleFunc("/", fghttp.ServeHttp)
		log.Fatal(http.ListenAndServe(ssocket, nil))
	}
}
