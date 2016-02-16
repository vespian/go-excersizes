// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Mandelbrot emits a PNG image of the Mandelbrot fractal.
package http

import (
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/vespian/go-excersizes/fractal_graphing/algos"
	"github.com/vespian/go-excersizes/fractal_graphing/common"
	"github.com/vespian/go-excersizes/fractal_graphing/img"
)

func ServeHttp(w http.ResponseWriter, r *http.Request) {
	width, height, scaling, algo, r_err := getHttpReqData(r)

	if r_err != nil {
		w.Header().Set("Content-Type", "text/plain")
		msg := fmt.Sprintf("getHttpReqData failed: %v", r_err)
		// (sur) How paranoid should I be ? do I need to do error checking
		// for each and every command or just common sense ?
		io.WriteString(w, msg)
		return
	}
	w.Header().Set("Content-Type", "img/png")
	f := algos.Str2funcMapping[algo]
	img := img.BuildImg(width, height, scaling, f)

	err := png.Encode(w, img) // NOTE: ignoring errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err while sending image to client: %v\n", err)
	}
}

func getHttpReqData(r *http.Request) (width, height, scaling int,
	algo string, r_err error) {
	var err error

	width, height, scaling, algo, err = parseHttpReqdata(r)
	if err != nil {
		r_err = fmt.Errorf("parseHttpReqdata failed: %v", err)
		return
	}

	err = common.ValidateImgParams(width, height, scaling, algo)
	if err != nil {
		r_err = fmt.Errorf("ValidateImgParams failed: %v", err)
		return
	}

	return
}

func parseHttpReqdata(r *http.Request) (width, height, scaling int,
	algo string, r_err error) {

	width, height = common.DEFAULT_WIDTH, common.DEFAULT_HEIGHT
	scaling = common.DEFAULT_SCALING
	algo = common.DEFAULT_ALGO

	var ok bool
	var tmp []string
	var err error

	if err = r.ParseForm(); err != nil {
		r_err = fmt.Errorf("Form parsing error: %v", err)
		return
	}

	//(sur) Is there simpler/more idiomatic way to parse headers, or should
	// I just parse them one by one and check the error status ?
	if tmp, ok = r.Form["width"]; ok {
		width, err = strconv.Atoi(tmp[0])
		if err != nil {
			r_err = fmt.Errorf("problem while parsing width to int: %v", err)
			return
		}
	}
	if tmp, ok = r.Form["height"]; ok {
		height, err = strconv.Atoi(tmp[0])
		if err != nil {
			r_err = fmt.Errorf("problem while parsing height to int: %v", err)
			return
		}
	}
	if tmp, ok = r.Form["scaling"]; ok {
		scaling, err = strconv.Atoi(tmp[0])
		if err != nil {
			r_err = fmt.Errorf("problem while parsing scaling to int: %v", err)
			return
		}
	}
	if tmp, ok = r.Form["algo"]; ok {
		algo = tmp[0]
	}

	return
}
