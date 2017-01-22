package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//{
//
//    "Title": "One Flew Over the Cuckoo's Nest",
//    "Year": "1975",
//    "Rated": "R",
//    "Released": "19 Nov 1975",
//    "Runtime": "133 min",
//    "Genre": "Drama",
//    "Director": "Milos Forman",
//    "Writer": "Lawrence Hauben (screenplay), Bo Goldman (screenplay), Ken Kesey (based on the novel by), Dale Wasserman (the play version: \"One Flew Over the Cuckoo's Nest\" by)",
//    "Actors": "Michael Berryman, Peter Brocco, Dean R. Brooks, Alonzo Brown",
//    "Plot": "A criminal pleads insanity after getting into trouble again and once in the mental institution rebels against the oppressive nurse and rallies up the scared patients.",
//    "Language": "English",
//    "Country": "USA",
//    "Awards": "Won 5 Oscars. Another 30 wins & 13 nominations.",
//    "Poster": "https://images-na.ssl-images-amazon.com/images/M/MV5BZjA0OWVhOTAtYWQxNi00YzNhLWI4ZjYtNjFjZTEyYjJlNDVlL2ltYWdlL2ltYWdlXkEyXkFqcGdeQXVyMTQxNzMzNDI@._V1_SX300.jpg",
//    "Metascore": "79",
//    "imdbRating": "8.7",
//    "imdbVotes": "701,658",
//    "imdbID": "tt0073486",
//    "Type": "movie",
//    "Response": "True"
//
//}

type movieData struct {
	title    string
	poster   string
	response bool
	err      string
}

func (md *movieData) UnmarshalJSON(in []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(in, &rawStrings)
	if err != nil {
		return err
	}

	for k, v := range rawStrings {
		switch strings.ToLower(k) {
		case "response":
			switch strings.ToLower(v) {
			case "true":
				md.response = true
			case "False":
				md.response = false
			default:
				return fmt.Errorf("Malformed `Response` value: `%s`", v)
			}
		case "title":
			md.title = v
		case "poster":
			md.poster = v
		case "error":
			md.err = v
		default:
			fmt.Printf("Skipping field `%s`\n", k)
		}
	}

	return nil
}

func assembleURL(movieName, apiURL string) (string, error) {
	// http://www.omdbapi.com/?t=nest&y=&plot=short&r=json
	res, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse omdb API url: %s", err)
	}

	params := url.Values{}
	params.Add("t", movieName)
	params.Add("y", "")
	params.Add("plot", "short")
	params.Add("r", "json")
	res.RawQuery = params.Encode()

	fmt.Printf("Encoded URL is %q\n", res.String())

	return res.String(), nil
}

func fetchDescription(movieName, apiURL string) (*movieData, error) {
	var err error
	var mData movieData

	url, err := assembleURL(movieName, apiURL)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("movie data json failed: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&mData); err != nil {
		return nil, fmt.Errorf("decoding json failed: %s", err)
	}

	fmt.Printf("Movie data is: `%+v`\n", mData)

	if !mData.response {
		return nil, fmt.Errorf("error from API: %s", mData.err)
	}

	return &mData, nil
}

func fetchPosterBlob(url string) ([]byte, error) {
	fmt.Printf("Fetching %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("movie data json failed: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func FetchPoster(movieName, apiURL string) ([]byte, error) {
	mData, err := fetchDescription(movieName, apiURL)
	if err != nil {
		return nil, err
	}

	blob, err := fetchPosterBlob(mData.poster)
	if err != nil {
		return nil, err
	}

	return blob, nil
}
