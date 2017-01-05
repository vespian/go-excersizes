package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type OndiskSerialization int

const (
	Protobuf OndiskSerialization = 1 + iota
	JSON
)

type OperationType int

const (
	Update OperationType = 1 + iota
	Search
	List
)

func (s OndiskSerialization) String() string {
	switch s {
	case Protobuf:
		return "protobuf"
	case JSON:
		return "json"
	default:
		return "unknown"
	}
}

func (s *OndiskSerialization) Set(in string) error {
	switch strings.ToLower(in) {
	case "protobuf":
		*s = Protobuf
	case "json":
		*s = JSON
	default:
		return fmt.Errorf("unrecognized serialization type `%s`", in)
	}
	return nil
}

func (s OperationType) String() string {
	switch s {
	case Update:
		return "update"
	case Search:
		return "search"
	case List:
		return "list"
	default:
		return "unknown"
	}
}

func (s *OperationType) Set(in string) error {
	switch strings.ToLower(in) {
	case "update":
		*s = Update
	case "search":
		*s = Search
	case "list":
		*s = List
	default:
		return fmt.Errorf("unrecognized operation `%s`", in)
	}
	return nil
}

// IndexFile bundles together all the parameters describing an index.
type IndexFile struct {
	Type     OndiskSerialization
	Location string
}

func (r IndexFile) String() string {
	return fmt.Sprintf("type: %s, location: %s", r.Type, r.Location)
}

type Range struct {
	Min, Max int
}

func (r Range) String() string {
	return fmt.Sprintf("min: %d, max: %d", r.Min, r.Max)
}

type Story struct {
	Alt        string
	Day        int `json:",string"`
	Img        string
	Link       string
	Month      int `json:",string"`
	News       string
	Num        int
	SafeTitle  string `json:"safe_title"`
	Title      string
	Transcript string
	Year       int `json:",string"`
}

func (s Story) String() string {
	res := ""

	res += fmt.Sprintf("\talt: %s\n", s.Alt)
	res += fmt.Sprintf("\tday: %d\n", s.Day)
	res += fmt.Sprintf("\timg: %s\n", s.Img)
	res += fmt.Sprintf("\tlink: %s\n", s.Link)
	res += fmt.Sprintf("\tmonth: %d\n", s.Month)
	res += fmt.Sprintf("\tnews: %s\n", s.News)
	res += fmt.Sprintf("\tnum: %d\n", s.Num)
	res += fmt.Sprintf("\tsafe_title: %s\n", s.SafeTitle)
	res += fmt.Sprintf("\ttitle: %s\n", s.Title)
	res += fmt.Sprintf("\ttranscript: %s\n", s.Transcript)
	res += fmt.Sprintf("\tyear: %d\n", s.Year)

	return res
}

type AllStories map[int]*Story

func (a AllStories) String() string {
	res := ""

	for i := range a {
		res += fmt.Sprintf("Story %d:\n%s", i, *a[i])
	}

	return res
}

func (a AllStories) MarshalJSON() ([]byte, error) {
	res := make(map[string]Story)

	for k, v := range a {
		res[strconv.Itoa(k)] = *v
	}

	return json.Marshal(res)
}

func (a AllStories) UnmarshalJSON(in []byte) error {
	res := make(map[string]Story)

	if err := json.Unmarshal(in, &res); err != nil {
		return err
	}

	for k, v := range res {
		var i int
		var err error
		if i, err = strconv.Atoi(k); err != nil {
			return err
		}
		a[i] = &v
	}

	return nil
}
