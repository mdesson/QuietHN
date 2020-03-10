package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const apiBase = "https://hacker-news.firebaseio.com/v0/"

// TODO: Create nonstatic template
var template = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Quiet Hacker News</title>
</head>
<body>
	<h1>Quiet Hacker News</h1>
<ol>
	<li><a href="https://www.google.ca">Literally just google.com</a> <span class="source">google.ca</span></li>
	<li><a href="https://www.yahoo.ca">Yahoo?</a> <span class="source">yahoo.ca</span></li>
</body>
</html>
`

// Contains a hackernews story
type story struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// Handler function, renders our only template
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, template)
}

// Returns an array integers, representing a story
func fetchTopStories() []int {
	url := apiBase + "topstories.json"
	var output []int

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err = json.Unmarshal(body, &output); err != nil {
		log.Fatal(err)
	}

	return output
}

// TODO: Given id, fetch story
func fetchStory(id int) {
	idString := strconv.Itoa(id)
	url := apiBase + "item/" + idString + ".json"
	var output story

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err = json.Unmarshal(body, &output); err != nil {
		log.Fatal(err)
	}

	fmt.Println(output.Title)
}

func main() {
	http.HandleFunc("/", handler)
	// log.Fatal(http.ListenAndServe("localhost:8000", nil))
	// fmt.Println("Started server on port 8000")
	id := fetchTopStories()[0]
	fetchStory(id)
}
