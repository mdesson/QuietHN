package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const apiBase = "https://hacker-news.firebaseio.com/v0/"

// TODO: Style page
var HNtemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Quiet Hacker News</title>
</head>
<body>
	<h1>Quiet Hacker News</h1>
<ol>
	{{range .}}
	<li><a href="{{.URL}}">{{.Title}}</a> <span class="source">{{.Domain}}</span>
	{{end}}
</ol>
</body>
</html>
`

// Contains a hackernews story
type story struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	Domain string
}

// Handler function, renders our only template
func handler(w http.ResponseWriter, r *http.Request) {
	stories := fetchTopThirty()
	parsedTemplate := template.Must(template.New("QuietHN").Parse(HNtemplate))
	parsedTemplate.Execute(w, stories)
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

// Given id, fetch story
func fetchStory(id int) story {
	idString := strconv.Itoa(id)
	storyUrl := apiBase + "item/" + idString + ".json"
	var output story

	res, err := http.Get(storyUrl)
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

	u, err := url.Parse(output.URL)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: Remove check on URL length once all selfposts are removed
	if len(u.Host) > 4 && u.Host[:4] == "www." {
		output.Domain = u.Host[4:]
	} else {
		output.Domain = u.Host
	}

	return output
}

// TODO: Fetch only links
func fetchTopThirty() []story {
	output := make([]story, 30)
	var wg sync.WaitGroup
	ids := fetchTopStories()
	for i, id := range ids[:30] {
		wg.Add(1)
		id := id
		i := i
		go func() {
			defer wg.Done()
			output[i] = fetchStory(id)
		}()
	}
	wg.Wait()
	return output
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server on port 8000")
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
