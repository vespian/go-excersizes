// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Package web groups all the code directly related to serving mandelbrot
// fractals through HTTP.
package web

import (
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/vespian/go-exercises/fractals/cmdline"
	"github.com/vespian/go-exercises/fractals/constants"
	"github.com/vespian/go-exercises/fractals/img"
)

// ServeHTTP responds with a PNG image of a fractal to the client.
func ServeHTTP(rW http.ResponseWriter, r *http.Request) {
	var err error

	imgP, err := getHTTPReqData(r)
	if err != nil {
		rW.Header().Set("Content-Type", "text/plain")
		msg := fmt.Sprintf("getHTTPReqData failed: %v", err)
		if _, err = io.WriteString(rW, msg); err != nil {
			msg = msg + ", and io.WriteString failed as well (???)"
			fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		}
		return
	}
	rW.Header().Set("Content-Type", "img/png")
	imgRes := img.BuildImg(imgP)

	err = png.Encode(rW, imgRes) // NOTE: ignoring errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err while sending image to client: %v\n", err)
	}
}

// getHTTPReqData - I have no idea how to name this function without get... :/
func getHTTPReqData(r *http.Request) (*cmdline.ImgParams, error) {
	var err error
	var imgP *cmdline.ImgParams

	imgP, err = parseHTTPReqdata(r)
	if err != nil {
		return nil, fmt.Errorf("parseHTTPReqdata failed: %v", err)
	}

	err = cmdline.ValidateImgParams(imgP)
	if err != nil {
		return nil, fmt.Errorf("ValidateImgParams failed: %v", err)
	}

	return imgP, nil
}

func parseHTTPReqdata(r *http.Request) (*cmdline.ImgParams, error) {
	var ok bool
	var tmp []string
	var err error

	imgP := cmdline.ImgParams{
		Width:   constants.DefaultWidth,
		Height:  constants.DefaultHeight,
		Scaling: constants.DefaultScaling,
		Algo:    constants.DefaultAlgo,
	}

	if err = r.ParseForm(); err != nil {
		rErr := fmt.Errorf("Form parsing error: %v", err)
		return nil, rErr
	}

	//(sur) Is there simpler/more idiomatic way to parse headers, or should
	// I just parse them one by one and check the error status ?
	if tmp, ok = r.Form["width"]; ok {
		imgP.Width, err = strconv.Atoi(tmp[0])
		if err != nil {
			rErr := fmt.Errorf("problem while parsing width to int: %v", err)
			return nil, rErr
		}
	}
	if tmp, ok = r.Form["height"]; ok {
		imgP.Height, err = strconv.Atoi(tmp[0])
		if err != nil {
			rErr := fmt.Errorf("problem while parsing height to int: %v", err)
			return nil, rErr
		}
	}
	if tmp, ok = r.Form["scaling"]; ok {
		imgP.Scaling, err = strconv.Atoi(tmp[0])
		if err != nil {
			rErr := fmt.Errorf("problem while parsing scaling to int: %v", err)
			return nil, rErr
		}
	}
	if tmp, ok = r.Form["algo"]; ok {
		imgP.Algo = tmp[0]
	}

	return &imgP, nil
}
