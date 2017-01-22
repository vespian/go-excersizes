package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/vespian/go-exercises/omdbapi/pkg/cmdline"
	"github.com/vespian/go-exercises/omdbapi/pkg/web"
)

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Operation failed: %s\n", err)
		os.Exit(1)
	}
}

func doMain() error {
	var err error

	args, err := cmdline.Get()

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", args)

	blob, err := web.FetchPoster(args.MovieName, args.OmdbapiURI)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(args.OutputFile, blob, os.FileMode(0660))

	return err
}
