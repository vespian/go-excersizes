// Package cmdline gathers together cmdline stuff used by xkcd app.
package cmdline

import (
	"flag"
	"fmt"
	"unsafe"

	"github.com/vespian/go-exercises/xkcd/pkg/types"
)

// CommandlineArgs bundles together all arguments that can be passed on the
// commandline.
type CommandlineArgs struct {
	types.IndexFile
	types.Range

	QueryString string
	Op          types.OperationType
	XkcdURI     string
}

func (c CommandlineArgs) String() string {
	var res string

	res += fmt.Sprintf("Commandline options:\n")
	res += fmt.Sprintf("  IndexFile: `%s`\n", c.IndexFile)
	res += fmt.Sprintf("  Range: `%s`\n", c.Range)
	res += fmt.Sprintf("  Query string: `%s`\n", c.QueryString)
	res += fmt.Sprintf("  XKCD uri: `%s`\n", c.XkcdURI)
	res += fmt.Sprintf("  Op: `%s`\n", c.Op)

	return res
}

func Parse() *CommandlineArgs {
	res := CommandlineArgs{
		Op:        types.List,
		IndexFile: types.IndexFile{Type: types.JSON},
	}

	flag.Var(&res.Type, "idx-type", "format of the on-disk index")
	flag.StringVar(&res.Location, "idx-file", "xkcd-index",
		"location of the offline index file (without extension)")
	flag.IntVar(&res.Min, "min", 1, "minimum xkcd index to process")
	flag.IntVar(&res.Max, "max", 1<<(unsafe.Sizeof(res.Max)*8-1)-1, "maximum xkcd index to process")
	flag.StringVar(&res.QueryString, "query", "", "String to use for filtering comics titles")
	flag.StringVar(&res.XkcdURI, "xkcd-uri-fmt", "http://xkcd.com/%d/info.0.json",
		"api endpoint address")
	flag.Var(&res.Op, "op", "operation to perform")

	flag.Parse()

	return &res
}
