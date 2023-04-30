package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/benjamineskola/rss2dayone/cache"
	"github.com/mmcdole/gofeed"
)

func main() { //nolint:cyclop
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <url> <journal> [tag...]\n", os.Args[0])
		os.Exit(1)
	}

	fp := gofeed.NewParser()

	feedURL := os.Args[1]

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		log.Fatal(err)
	}

	downloadDir, err := os.MkdirTemp("", "rss2dayone-")
	if err != nil {
		log.Fatalf("Could not create temporary directory: %s", err)
	}

	defer os.Remove(downloadDir)

	processedAny := false

	processed, err := cache.Init()
	if err != nil {
		log.Panic("Could not load seen items cache: ", err)
	}

	for _, item := range feed.Items {
		if processed.Contains(item.GUID) {
			continue
		}

		post, err := NewPost(item, downloadDir)
		if err != nil {
			log.Print(err)

			continue
		}

		if err := invokeDayOne(post, os.Args[2], os.Args[3:]); err != nil {
			log.Printf("failed invocation of dayone2: %s", err)

			continue
		}

		processed.Add(item.GUID)

		processedAny = true
	}

	if processedAny {
		if err = processed.Save(); err != nil {
			log.Panic("Failed to save seen items cache: ", err)
		}
	}
}

func handleLetterboxdExtensions(item *gofeed.Item, title string, postTime time.Time) (string, time.Time, error) {
	if len(item.Extensions["letterboxd"]["filmTitle"]) > 0 &&
		len(item.Extensions["letterboxd"]["filmYear"]) > 0 {
		title = fmt.Sprintf("%s (%s)",
			item.Extensions["letterboxd"]["filmTitle"][0].Value,
			item.Extensions["letterboxd"]["filmYear"][0].Value)
	}

	if len(item.Extensions["letterboxd"]["watchedDate"]) > 0 {
		var err error

		postTime, err = time.Parse("2006-01-02", item.Extensions["letterboxd"]["watchedDate"][0].Value)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("could not parse time of %s: %w", item.GUID, err)
		}
	}

	return title, postTime, nil
}

func invokeDayOne(post *Post, journal string, tags []string) error {
	cmdArgs := []string{"new", "--journal", journal, "--isoDate", post.date.Format("2006-01-02T15:04:05"), "--tags"}
	cmdArgs = append(cmdArgs, tags...)

	for _, i := range *post.AttachmentFiles {
		cmdArgs = append(cmdArgs, "-a", i)

		defer os.Remove(i)
	}

	body := ""
	if len(post.title) > 0 {
		body += "# " + post.title + "\n"
	}

	if len(*post.AttachmentFiles) > 0 {
		body += "[{attachment}]\n"
	}

	body += post.body

	cmd := exec.Command("dayone2", cmdArgs...)
	cmd.Stdin = strings.NewReader(body)

	var out strings.Builder
	cmd.Stdout = &out

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute dayone2: %w", err)
	}

	log.Print(out.String())
	log.Print(stderr.String())

	return nil
}
