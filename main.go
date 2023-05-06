package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/benjamineskola/rss2dayone/cache"
	"github.com/mmcdole/gofeed"
)

func main() {
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

	processed, err := cache.Init()
	if err != nil {
		log.Panic("Could not load seen items cache: ", err)
	}

	for _, item := range feed.Items {
		if processed.Contains(item.GUID) {
			continue
		}

		post, err := NewPost(item)
		if err != nil {
			log.Print(err)

			continue
		}

		if err := invokeDayOne(post, os.Args[2], os.Args[3:]); err != nil {
			log.Printf("failed invocation of dayone2: %s", err)

			continue
		}

		processed.Add(item.GUID)
	}

	if err = processed.Save(); err != nil {
		log.Panic("Failed to save seen items cache: ", err)
	}
}

func invokeDayOne(post *Post, journal string, tags []string) error {
	cmdArgs := []string{"new", "--journal", journal, "--isoDate", post.date.Format("2006-01-02T15:04:05"), "--tags"}
	cmdArgs = append(cmdArgs, tags...)

	downloadDir, err := os.MkdirTemp("", "rss2dayone-")
	if err != nil {
		log.Fatalf("Could not create temporary directory: %s", err)
	}

	defer os.Remove(downloadDir)

	post.FetchAttachments(downloadDir)

	for _, i := range *post.AttachmentFiles {
		cmdArgs = append(cmdArgs, "-a", i)

		defer os.Remove(i)
	}

	cmd := exec.Command("dayone2", cmdArgs...)
	cmd.Stdin = strings.NewReader(post.Render())

	var stdout strings.Builder
	cmd.Stdout = &stdout

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute dayone2: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}
