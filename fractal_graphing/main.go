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
	"io"
	"log"
	"math/cmplx"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
)

const XMIN, YMIN, XMAX, YMAX = -2.2, -1.1, +2.2, +1.1
const TILE_SIZE = 64
const DEFAULT_WIDTH, DEFAULT_HEIGHT = 2048, 1024
const DEFAULT_SCALING = 1
const DEFAULT_ALGO = "mandelbrot_c128"

type algo_func func(r, i float64) (uint8, uint8)

// (sur) Is there a way to do _getattr_ in a simple way ? Or is the "map"
// approach idiomatic enough ?
var str2func_mapping = map[string]algo_func{
	"newton":          newton,
	"acos":            acos,
	"mandelbrot_c64":  mandelbrot_c64,
	"mandelbrot_c128": mandelbrot_c128,
	"sqrt":            sqrt,
}

// TODO:
// - split into separate files

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	width, height, scaling, filepath, ssocket, algo := get_cmdline()

	if filepath != "" {
		f := str2func_mapping[algo]
		img := build_img(width, height, scaling, f)
		// (sur) If we close a buffered channel with messages in the buffer,
		// do we lose them or are they going to be consumed by the other end and
		// then it notices that ch is closed ?
		write_img(img, filepath)
	}

	if ssocket != "" {
		http.HandleFunc("/", serve_http)
		log.Fatal(http.ListenAndServe(ssocket, nil))
	}

}

func get_http_reqdata(r *http.Request) (width, height, scaling int,
	algo string, r_err error) {
	var err error

	width, height, scaling, algo, err = parse_http_reqdata(r)
	if err != nil {
		r_err = fmt.Errorf("parse_http_reqdata failed: %v", err)
		return
	}

	err = validate_img_params(width, height, scaling, algo)
	if err != nil {
		r_err = fmt.Errorf("validate_img_params failed: %v", err)
		return
	}

	return
}

func parse_http_reqdata(r *http.Request) (width, height, scaling int,
	algo string, r_err error) {

	width, height = DEFAULT_WIDTH, DEFAULT_HEIGHT
	scaling = DEFAULT_SCALING
	algo = DEFAULT_ALGO

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

func serve_http(w http.ResponseWriter, r *http.Request) {
	width, height, scaling, algo, r_err := get_http_reqdata(r)

	if r_err != nil {
		w.Header().Set("Content-Type", "text/plain")
		msg := fmt.Sprintf("get_http_reqdata failed: %v", r_err)
		// (sur)How paranoid we should be ? do I need to do error checking here
		// as well ?
		io.WriteString(w, msg)
		return
	}
	w.Header().Set("Content-Type", "img/png")
	f := str2func_mapping[algo]
	img := build_img(width, height, scaling, f)

	err := png.Encode(w, img) // NOTE: ignoring errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err while sending image to client: %v\n", err)
	}

}

func parse_cmdline() (width, height, scaling int, filepath, ssocket, algo string) {
	flag.IntVar(&width, "width", DEFAULT_WIDTH, "Width of the resulting image")
	flag.IntVar(&height, "height", DEFAULT_HEIGHT, "Height of the resulting image")
	flag.IntVar(&scaling, "scaling", DEFAULT_SCALING,
		"Super-sampling factor of the image (1 == no super-sampling)")
	flag.StringVar(&filepath, "filepath", "",
		"File, where resulting image is going to be saved")
	flag.StringVar(&ssocket, "ssocket", "",
		"Http server socket")
	//(sur) Is this the idiomatic way how to break/format strings ?
	flag.StringVar(&algo, "algorithm", DEFAULT_ALGO,
		"Algorithm to use to calculate the fractal (newton|acos|mandelbrot"+
			"(_c64|_c128)|sqrt)")

	flag.Parse()

	return
}

func validate_img_params(width, height, scaling int, algo string) error {
	if width%TILE_SIZE != 0 || height%TILE_SIZE != 0 {
		msg_fmt := "Width(%d) and height(%d) of the resulting picture" +
			" should be multiples of tile size(%d)\n"
		return fmt.Errorf(msg_fmt, width, height, TILE_SIZE)
	}

	ratio_xy := float64(YMAX-YMIN) / float64(XMAX-XMIN)
	ratio_pxpy := float64(height) / float64(width)
	if ratio_xy != ratio_pxpy {
		sugested_width := int(float64(height) / ratio_xy)
		msg_fmt := "Pixel ratio (%2.2f) differs from XY ratio(%2.2f), try" +
			" adjusting width to %d\n\n"
		return fmt.Errorf(msg_fmt, ratio_pxpy, ratio_xy, sugested_width)
	}

	if scaling < 1 {
		msg_fmt := "Scaling factor must be >= 1, currently: `%d`\n"
		return fmt.Errorf(msg_fmt, scaling)
	}

	if _, ok := str2func_mapping[algo]; !ok {
		msg := "algorithm must be one of: mandelbrot_(c64|c128)|sqrt|acos" +
			"|newton, given: %s\n"
		return fmt.Errorf(msg, algo)
	}

	return nil

}

func validate_cmdline(width, height, scaling int, filepath, ssocket, algo string) {
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
		err := validate_img_params(width, height, scaling, algo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
	} else {
		// Validated while handling HTTP request
	}

}

func get_cmdline() (width, height, scaling int, filepath, ssocket, algo string) {
	width, height, scaling, filepath, ssocket, algo = parse_cmdline()
	validate_cmdline(width, height, scaling, filepath, ssocket, algo)
	return
}

// (sur) One more question about breaking the line (here a func. signature) :)
func build_img(width, height, scaling int,
	algo func(r, i float64) (uint8, uint8)) *image.RGBA {

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	wait_group, worker_ch := spawn_processors(scaling, img, algo)

	for py := 0; py < height; py += TILE_SIZE {
		for px := 0; px < width; px += TILE_SIZE {
			// (sur) should I worry about generating to many objects for GC ?
			worker_ch <- [2]int{py, px}
		}
	}

	close(worker_ch)
	wait_group.Wait()

	return img
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

func spawn_processors(subsample_f int, img *image.RGBA, algo algo_func) (
	*sync.WaitGroup, chan<- [2]int) {
	// (sur) Should we worry about GC in general if we are sending too many
	// small elements through channel ?
	num_threads := runtime.NumCPU()
	worker_ch := make(chan [2]int, num_threads*2)
	var wg sync.WaitGroup

	for i := 0; i < num_threads; i++ {
		wg.Add(1)
		go tile_calculator(subsample_f, img, worker_ch, &wg, i, algo)
	}

	return &wg, worker_ch
}

//(sur) Is this the canonical way to break this line ?
func tile_calculator(subsample_f int,
	img *image.RGBA,
	in <-chan [2]int,
	wg *sync.WaitGroup, thread_num int,
	algo algo_func) {

	fmt.Fprintf(os.Stderr, "goroutine %d starting\n", thread_num)

	img_bounds := img.Bounds()
	y_delta := (YMAX - YMIN) / float64(img_bounds.Max.Y)
	x_delta := (XMAX - XMIN) / float64(img_bounds.Max.X)

	for d := range in {
		// (sur) What is the idiomatic way to unpack a list into vars ?
		// Something like: ?
		py, px := d[0], d[1]
		//fmt.Fprintf(os.Stderr, "Processing tile ((`%d`,`%d`),(`%d`,`%d`))\n",
		//py, px, py+TILE_SIZE, px+TILE_SIZE)
		var offset float64 = 0
		if subsample_f > 1 {
			offset = 0.5
		}

		y_base := (float64(py)-offset)*y_delta + YMIN
		x_base := (float64(px)-offset)*x_delta + XMIN

		for i_y := 0; i_y < TILE_SIZE; i_y++ {
			for i_x := 0; i_x < TILE_SIZE; i_x++ {
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

// Mandelbrot algo, but with complex128 type resolution
func mandelbrot_c128(r, i float64) (uint8, uint8) {
	var z complex128 = complex(r, i)
	const iterations = 20000

	var v complex128
	for n := 0; n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			blue := uint8(real(v)*128) + 127
			red := uint8(imag(v)*128) + 127
			return blue, red
		}
	}
	return 0, 0
}

// Mandelbrot algo, but with complex64 type resolution
func mandelbrot_c64(r, i float64) (uint8, uint8) {
	var z complex64 = complex(float32(r), float32(i))
	const iterations = 20000

	var v complex64
	for n := 0; n < iterations; n++ {
		v = v*v + z
		//(sur) I have not found a better way, golang seems to support only
		// float64 arithmetics :/
		if float32(cmplx.Abs(complex128(v))) > 2 {
			blue := uint8(real(v)*128) + 127
			red := uint8(imag(v)*128) + 127
			return blue, red
		}
	}
	return 0, 0
}

func acos(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Acos(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

func sqrt(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	v := cmplx.Sqrt(z)
	blue := uint8(real(v)*128) + 127
	red := uint8(imag(v)*128) + 127
	return blue, red
}

// f(x) = x^4 - 1
//
// z' = z - f(z)/f'(z)
//    = z - (z^4 - 1) / (4 * z^3)
//    = z - (z - 1/z^3) / 4
func newton(r, i float64) (uint8, uint8) {
	z := complex(r, i)
	const iterations = 20000

	for i := 0; i < iterations; i++ {
		z -= (z - 1/(z*z*z)) / 4
		if cmplx.Abs(z*z*z*z-1) < 1e-6 {
			blue := uint8(real(z)*128) + 127
			red := uint8(imag(z)*128) + 127
			return blue, red
		}
	}

	return 0, 0
}
