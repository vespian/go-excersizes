#fractal_graphing#
An excercise from "Go Programming Language" book: p. 62, ex.3.5-3.9

The program itself generates images of few simple equations/fractals (check
/algos/algos.go file). Depending on command line it either produces output to
a file, or starts a web-server serving PNG files over the HTTP.

## Operation ##
Depending on the command-line parameters (for cmdline operation) or HTTP GET
parameters (for WWW server operation) it produces image that has following
attributes:
- width, size
- scaling: supersamping ratio of the resulting image (1 pixel of the resulting
  image == average of  X point calculated by program, where X == scaling param)
- algo: algorithm to use for calculating the image (check help for a full list)

## Comments/questions ##
Sergiusz - in the code I have made comments marked with (sur) tag - could you
please take  a look and comment on them ?
