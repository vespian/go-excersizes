package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vespian/go-exercises/xkcd/pkg/types"
)

func fetchStory(url string) (*types.Story, error) {
	var result types.Story

	fmt.Printf("Fetching url %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching story %s failed: %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding story %s failed: %s", url, err)
	}

	return result, nil
}

func Fetch(format string, rg types.Range) (types.AllStories, error) {
	res := make(types.AllStories)

	for i := rg.Min; i < rg.Max; i++ {
		url := fmt.Sprintf(format, i)
		story, err := fetchStory(url)
		if err != nil {
			return nil, fmt.Errorf("Fetch failed: %s", err)
		}
		if story == nil {
			fmt.Printf("Reached end of stories at id %d\n", i)
			break
		}
		fmt.Printf("Fetched story %d\n", i)
		res[i] = story
	}

	return res, nil
}
