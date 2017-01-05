package index

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/vespian/go-exercises/xkcd/pkg/pbuff"
	"github.com/vespian/go-exercises/xkcd/pkg/types"
	"github.com/vespian/go-exercises/xkcd/pkg/web"
)

func serialize(a types.AllStories, t types.OndiskSerialization,
) (
	[]byte,
	error,
) {
	var blob []byte
	var err error

	switch t {
	case types.JSON:
		blob, err = json.Marshal(a)
		if err != nil {
			return nil, fmt.Errorf("Index JSON marshaling failed: %s", err)
		}
	case types.Protobuf:
		pbuffDigestableStructs := pbuff.PBAllStoriesFromAllStories(a)
		blob, err = proto.Marshal(pbuffDigestableStructs)
		if err != nil {
			log.Fatal("pbuff marshaling error: ", err)
		}
	default:
		panic("Unsupported serializing method")
	}

	return blob, err
}

func deserialize(in []byte, t types.OndiskSerialization,
) (
	types.AllStories,
	error,
) {
	var res types.AllStories
	var err error

	switch t {
	case types.JSON:
		if err = json.Unmarshal(in, &res); err != nil {
			log.Fatalf("JSON unmarshaling failed: %s", err)
		}
	case types.Protobuf:
		tmp := new(pbuff.PBAllStories)

		err = proto.Unmarshal(in, tmp)
		if err != nil {
			log.Fatal("pbuff unmarshaling error: ", err)
		}
		res = pbuff.AllStoriesFromPBAllStories(tmp)
	default:
		panic("Unsupported serializing method")
	}

	return res, err
}

func read(idx types.IndexFile) (types.AllStories, error) {
	var blob []byte
	var a types.AllStories
	var err error

	if blob, err = ioutil.ReadFile(idx.Location); err != nil {
		return nil, err
	}

	if a, err = deserialize(blob, idx.Type); err != nil {
		return nil, err
	}

	return a, nil
}

func store(a types.AllStories, idx types.IndexFile) error {
	var err error
	var blob []byte

	if blob, err = serialize(a, idx.Type); err != nil {
		return err
	}

	if err = ioutil.WriteFile(idx.Location, blob, os.FileMode(0540)); err != nil {
		return err
	}

	return nil
}

func filterByRange(a types.AllStories, rg types.Range) types.AllStories {
	res := types.AllStories{}

	for k := range a {
		if k >= rg.Min && k <= rg.Max {
			res[k] = a[k]
		}
	}

	return res
}

func filterByQuery(a types.AllStories, query string) types.AllStories {
	res := types.AllStories{}

	for k, v := range a {
		if strings.Contains(v.Title, query) {
			res[k] = a[k]
		}
	}

	return res
}

func Update(url string, rg types.Range, idx types.IndexFile) error {
	var a types.AllStories
	var err error

	fmt.Printf("Fetching from `%s`, range: `%s`, idx: `%s`\n", url, rg, idx)

	if a, err = web.Fetch(url, rg); err != nil {
		return err
	}
	if err = store(a, idx); err != nil {
		return err
	}

	return nil
}

func Fetch(query string, rg types.Range, idx types.IndexFile) (types.AllStories, error) {
	var a types.AllStories
	var err error

	fmt.Printf("Listing entries, "+
		"query: `%s`, "+
		"range: `%s`, "+
		"idx: `%s`\n", query, rg, idx)

	if a, err = read(idx); err != nil {
		return nil, err
	}

	if query != "" {
		a = filterByQuery(a, query)
	}

	if rg.Min > 0 || rg.Max > 0 {
		a = filterByRange(a, rg)
	}

	return a, nil
}
