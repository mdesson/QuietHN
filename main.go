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
	"time"
)

// TODO: Keep cache state in main()
// TODO: Set cache to update on ticker

const apiBase = "https://hacker-news.firebaseio.com/v0/"

var templateHN = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Quiet Hacker News</title>
	<style>
		body {
			font-family: Helvetica, Arial, sans-serif;
		}
		li {
			margin: 0.5em
		}
		a {
			color: black;
		}
		.source {
			color: grey;
		}
	</style>
</head>
<body>
	<h1>Quiet Hacker News</h1>
<ol>
	{{range .}}
	<li><a href="{{.URL}}">{{.Title}}</a> <span class="source">({{.Domain}})</span>
	{{end}}
</ol>
</body>
</html>
`

// Story Contains a hackernews story
type Story struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	Domain string
}

// Handler function, renders our only template
func handler(w http.ResponseWriter, r *http.Request) {
	stories := fetchTopThirty()
	parsedTemplate := template.Must(template.New("QuietHN").Parse(templateHN))
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
func fetchStory(id int) Story {
	idString := strconv.Itoa(id)
	storyURL := apiBase + "item/" + idString + ".json"
	var output Story

	res, err := http.Get(storyURL)
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

	if len(u.Host) > 4 && u.Host[:4] == "www." {
		output.Domain = u.Host[4:]
	} else {
		output.Domain = u.Host
	}

	return output
}

// Fetches top 30 stories
func fetchTopThirty() []Story {
	fmt.Print("Fetching stories... ")
	start := time.Now()

	output := make([]Story, 30)
	ids := fetchTopStories()[:120]
	stories := make([]Story, 120)
	outputFull := make(chan bool)
	doneFetching := make(chan bool)
	readyToParse := make(chan bool)

	// Parse goroutine: Parses through stories, filling output
	go func() {
		inputIndex := 0
		outputIndex := 0

		// Loop over each batch of 40
		for i := 0; i < 3; i++ {
			// Block until batch is ready
			<-readyToParse
			// Fill as much of output as possible during batch
			for outputIndex < 30 {
				for _, story := range stories[inputIndex : inputIndex+40] {
					if story.URL != "" && story.Type == "story" {
						output[outputIndex] = story
						inputIndex++
						break
					}
					inputIndex++
				}
				outputIndex++
			}
			inputIndex = i * 40

			// End loop if all 30 slots are filled
			if outputIndex >= 30 {
				outputFull <- true
				return
			}
		}
	}()

	// Fetch goroutines: Loop over 40/80/120 top stories
	for i := 0; i < 3; i++ {
		index := i * 40
		go func() {
			var wg sync.WaitGroup
			wg.Add(40)
			for i, id := range ids[index : index+40] {
				i, id := i, id
				go func() {
					defer wg.Done()
					stories[i] = fetchStory(id)
				}()
			}
			wg.Wait()
			// Signal to loop ready to continue
			doneFetching <- true
		}()

		select {
		// All 30 found, end loop
		case <-outputFull:
			elapsed := time.Now().Sub(start)
			fmt.Printf("Complete (%v)\n", elapsed)
			return output
		case <-doneFetching:
			readyToParse <- true
		}
	}
	fmt.Println("outputFull select not triggered")
	return output
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server on port 8000")
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
