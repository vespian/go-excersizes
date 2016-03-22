// Copyright © 2016 Pawel Rozlach.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Based on work by Alan A. A. Donovan & Brian W. Kernighan
// which can be found at:
// https://github.com/adonovan/gopl.io.git

// Package constants contains misc code used by both webserver and fileoutput
// variants.
package constants

// XMin - begining of X axis
const XMin = -2.2

// YMin - begining of Y axis
const YMin = -1.1

// XMax - end of x axis
const XMax = 2.2

// YMax - end of Y axis
const YMax = 1.1

// TileSize defines a basic unit of processing in pixels. Resulting image is
// composed of tiles of TileSizexTileSize size.
const TileSize = 64

// DefaultWidth is the number of pixels that can by mapped to given X axis
// range.
const DefaultWidth = 2048

// DefaultHeight is the number of pixels that can by mapped to given Y axis
// range.
const DefaultHeight = 1024

// DefaultScaling is default superscalling factor.
const DefaultScaling = 1

// DefaultAlgo is defaut algorithm to use
const DefaultAlgo = "mandelbrotC128"
