package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"github.com/benjamineskola/rss2dayone/cache"
	"github.com/mmcdole/gofeed"
)

type Feed struct {
	URL     string
	Journal string
	Tags    []string
}

type Config struct {
	Feeds []Feed `toml:"feed"`
}

func main() {
	configPath := filepath.Join(xdg.ConfigHome, "rss2dayone.toml")

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Panicf("cannot read config file %s: %s", configPath, err)
	}

	var config Config
	if _, err := toml.NewDecoder(configFile).Decode(&config); err != nil {
		log.Panicf("error decoding TOML: %s", err)
	}

	processed, err := cache.Init()
	if err != nil {
		log.Panic("Could not load seen items cache: ", err)
	}

	for _, feed := range config.Feeds {
		if err = processFeed(feed.URL, feed.Journal, feed.Tags, processed); err != nil {
			log.Print(err)
		}
	}

	if err = processed.Save(); err != nil {
		log.Panic("Failed to save seen items cache: ", err)
	}
}

func processFeed(feedURL, journal string, tags []string, processed *cache.Cache) error {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return fmt.Errorf("failed to process feed %s: %w", feedURL, err)
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

		if err := invokeDayOne(post, journal, tags); err != nil {
			log.Printf("failed invocation of dayone2: %s", err)

			continue
		}

		processed.Add(item.GUID)
	}

	return nil
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
