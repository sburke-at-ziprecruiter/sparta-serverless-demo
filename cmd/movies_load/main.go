package main

import (
	"MyGo/dynamo_db/pkg/movie"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"

	"fmt"
	"os"
)

// readMovies reads Movies from a gzipped JSON via stdin.
// Note that these movies need to call Init() to set PK and SK.
func readMovies() ([]movie.Movie, error) {
	gunzip, err := gzip.NewReader(os.Stdin)
	if err != nil {
		return nil, err
	}
	defer gunzip.Close()

	raw, err := ioutil.ReadAll(gunzip)
	if err != nil {
		return nil, err
	}

	var movies []movie.Movie
	err = json.Unmarshal(raw, &movies)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func main() {

	// Load Movies via stdin and add them the DDB table.
	movies, err := readMovies()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for ix, movie := range movies {
		if ix > 9 {
			break
		}
		if err = movie.Put(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	// Scan the movies that we loaded.
	movies, err = movie.Scan()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for _, movie := range movies {
		fmt.Printf("Successfully added %q (%d) to table.\n", movie.Title, movie.Year)
	}
}
