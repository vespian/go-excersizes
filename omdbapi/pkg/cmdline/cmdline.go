// Package cmdline gathers together cmdline stuff used by xkcd app.
package cmdline

import (
	"flag"
	"fmt"
)

// CommandlineArgs bundles together all arguments that can be passed on the
// command line.
type CommandlineArgs struct {
	MovieName  string
	OutputFile string
	OmdbapiURI string
}

func (c CommandlineArgs) String() string {
	var res string

	res += fmt.Sprintf("Commandline options:\n")
	res += fmt.Sprintf("  Movie name: `%s`\n", c.MovieName)
	res += fmt.Sprintf("  OmdbapiURI uri: `%s`\n", c.OmdbapiURI)
	res += fmt.Sprintf("  OutputFile: `%s`\n", c.OutputFile)

	return res
}

func Parse() *CommandlineArgs {
	res := CommandlineArgs{}

	flag.StringVar(&res.MovieName,
		"query",
		"",
		"Movie for which to download poster")
	flag.StringVar(&res.OmdbapiURI,
		"omdbapi-uri",
		"https://omdbapi.com/",
		"api endpoint address")
	flag.StringVar(&res.OutputFile,
		"output-file",
		"poster.jpg",
		"file to write output image to")

	flag.Parse()

	return &res
}

func Verify(cmd *CommandlineArgs) error {
	if cmd.MovieName == "" {
		return fmt.Errorf("Movie name is empty")
	}

	return nil
}

func Get() (*CommandlineArgs, error) {
	cmd := Parse()
	if err := Verify(cmd); err != nil {
		return nil, fmt.Errorf("Commandline verification failed: %s", err)
	}

	return cmd, nil
}
