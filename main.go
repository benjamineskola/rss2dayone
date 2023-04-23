package main

import (
	"log"
	"os"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()

	feedURL := os.Args[1]

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(feed.Title)

	for _, item := range feed.Items {
		log.Print(item.GUID)
		processItem(item)
	}
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
