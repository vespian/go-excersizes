// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Package cmdline gathers together cmdline stuff used by fractals
// app.
package cmdline

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/vespian/go-exercises/fractals/algos"
	"github.com/vespian/go-exercises/fractals/constants"
)

// ImgParams bundles together all parameters that are needed to construct an
// image.
type ImgParams struct {
	Width   int
	Height  int
	Scaling int
	Algo    string
}

// CommandlineArgs bundles together all arguments that can be passed on the
// commandline.
type CommandlineArgs struct {
	ImgParams
	Filepath string
	Ssocket  string
	Err      error
}

func parseCmdline() *CommandlineArgs {
	var res CommandlineArgs

	flag.IntVar(&res.Width, "width", constants.DefaultWidth,
		"Width of the resulting image")
	flag.IntVar(&res.Height, "height", constants.DefaultHeight,
		"Height of the resulting image")
	flag.IntVar(&res.Scaling, "scaling", constants.DefaultScaling,
		"Super-sampling factor of the image (1 == no super-sampling)")
	flag.StringVar(&res.Filepath, "filepath", "",
		"File, where resulting image is going to be saved")
	flag.StringVar(&res.Ssocket, "ssocket", "",
		"Http server socket")
	//(sur) Is this the idiomatic way how to break/format strings ?
	flag.StringVar(&res.Algo, "algorithm", constants.DefaultAlgo,
		"Algorithm to use to calculate the fractal (newton|acos|mandelbrot"+
			"(_c64|_c128)|sqrt)")

	flag.Parse()

	return &res
}

func validateCmdline(cmd *CommandlineArgs) {
	switch {
	case cmd.Filepath == "" && cmd.Ssocket == "":
		cmd.Err = fmt.Errorf("Either path to an output file or server socket " +
			"must be specified")
		return

	case cmd.Filepath != "" && cmd.Ssocket != "":
		cmd.Err = fmt.Errorf("Filepath and ssocket are mutually exclusive\n")
		return

	case cmd.Filepath != "":
		err := ValidateImgParams(&cmd.ImgParams)
		if err != nil {
			cmd.Err = fmt.Errorf("Img params are invalid: %s", err)
			return
		}
		// Image params validation for HTTP server occurs during req. processing
	}
}

// ValidateImgParams validates desired output image parameters.
func ValidateImgParams(imgP *ImgParams) error {
	if imgP.Width%constants.TileSize != 0 || imgP.Height%constants.TileSize != 0 {
		return fmt.Errorf("Width(%d) and height(%d) of the resulting picture"+
			" should be multiples of tile size(%d)\n", imgP.Width,
			imgP.Height, constants.TileSize)
	}

	ratioXY := float64(constants.YMax-constants.YMin) / float64(constants.XMax-constants.XMin)
	ratioPxPy := float64(imgP.Height) / float64(imgP.Width)
	if ratioXY != ratioPxPy {
		sugestedWidth := int(float64(imgP.Height) / ratioXY)
		msgFmt := "Pixel ratio (%2.2f) differs from XY ratio(%2.2f), try" +
			" adjusting width to %d\n\n"
		return fmt.Errorf(msgFmt, ratioPxPy, ratioXY, sugestedWidth)
	}

	if imgP.Scaling < 1 {
		msgFmt := "Scaling factor must be >= 1, currently: `%d`\n"
		return fmt.Errorf(msgFmt, imgP.Scaling)
	}

	if _, err := algos.MapStr2Func(imgP.Algo); err != nil {
		return err
	}

	return nil
}

// Cmdline parses and validates commandline arguments.
func Cmdline() *CommandlineArgs {
	cmd := parseCmdline()
	validateCmdline(cmd)
	return cmd
}

// WriteImg writes given image to a file.
func WriteImg(img *image.RGBA, filepath string) (err error) {
	var f *os.File

	if f, err = os.Create(filepath); err != nil {
		return errors.New("Couldn't open file: " + err.Error())
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	w := bufio.NewWriter(f)

	if err := png.Encode(w, img); err != nil {
		return err
	}

	return nil
}
