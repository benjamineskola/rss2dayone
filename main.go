package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/adrg/xdg"
	"github.com/mmcdole/gofeed"
)

var CACHEFILE string = filepath.Join(xdg.CacheHome, "rss2dayone.json")

func main() {
	fp := gofeed.NewParser()

	feedURL := os.Args[1]

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		log.Fatal(err)
	}

	processedAny := false

	processed := loadProcessedItemsList()
	for _, item := range feed.Items {
		_, isPresent := processed[item.GUID]
		if isPresent {
			continue
		}

		processItem(item)

		processed[item.GUID] = struct{}{}
		processedAny = true
	}

	if processedAny {
		saveProcessedItemsList(&processed)
	}
}

func loadProcessedItemsList() map[string]struct{} {
	// for compatibility with the python version this is saved as a list, but
	// the go code wants it as a map for efficiency: O(1) instead of O(n)
	processedList := make([]string, 0)

	data, err := os.ReadFile(CACHEFILE)
	if err == nil {
		err = json.Unmarshal(data, &processedList)
		if err != nil {
			log.Fatal("Error decoding processed map:", err)
		}
	}

	processed := make(map[string]struct{})
	for _, i := range processedList {
		// log.Print("seen: ", i)
		processed[i] = struct{}{}
	}

	return processed
}

func processItem(item *gofeed.Item) {
	log.Print("Title: ", item.Title)

	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(item.Description)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Description: %s\n", markdown)
	log.Printf("Link: %s\n", item.Link)
	log.Printf("Published: %s\n", item.Published)

	for n, ext := range item.Extensions["letterboxd"] {
		log.Printf("%s\t%s\n", n, ext[0].Value)
	}
}

func saveProcessedItemsList(processed *map[string]struct{}) {
	processedList := make([]string, 0)
	for i := range *processed {
		processedList = append(processedList, i)
	}

	data, err := json.MarshalIndent(processedList, "", "  ")
	if err != nil {
		log.Fatal("Error serialising seen data:", err)
	}

	err = os.WriteFile(CACHEFILE, data, 0o644)
	if err != nil {
		log.Fatal("Error writing seen data to file:", err)
	}
}
