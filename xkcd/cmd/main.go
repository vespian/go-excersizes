// xkcd is an xkcd offline indexing and quering tool.
package main

import (
	"fmt"
	"os"

	"github.com/vespian/go-exercises/xkcd/pkg/cmdline"
	"github.com/vespian/go-exercises/xkcd/pkg/index"
	"github.com/vespian/go-exercises/xkcd/pkg/types"
)

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Operation failed: %s\n", err)
		os.Exit(1)
	}
}

func doMain() error {
	var err error

	c := cmdline.Parse()
	fmt.Printf("%+v\n", c)

	switch c.Op {
	case types.Update:
		err = index.Update(c.XkcdURI, c.Range, c.IndexFile)
	case types.List:
		var d types.AllStories
		if d, err = index.Fetch("", c.Range, c.IndexFile); err == nil {
			fmt.Print(d)
		}
	case types.Search:
		var d types.AllStories
		if d, err = index.Fetch(c.QueryString, c.Range, c.IndexFile); err == nil {
			fmt.Println(d)
		}
	default:
		// Should not happen, but still, just in case:
		err = fmt.Errorf("Unsupported operation: `%s`\n", c.Op)
	}

	return err
}
