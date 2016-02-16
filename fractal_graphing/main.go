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

// TODO:
// - split into separate files

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	width, height, scaling, filepath, ssocket, algo := cmdline.GetCmdline()

	if filepath != "" {
		f := algos.Str2funcMapping[algo]
		img := img.BuildImg(width, height, scaling, f)
		// (sur) If we close a buffered channel with messages in the buffer,
		// do we lose them or are they going to be consumed by the other end and
		// then it notices that ch is closed ?
		cmdline.WriteImg(img, filepath)
	}

	if ssocket != "" {
		http.HandleFunc("/", fghttp.ServeHttp)
		log.Fatal(http.ListenAndServe(ssocket, nil))
	}
}
