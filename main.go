package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/benjamineskola/rss2dayone/cache"
	"github.com/mmcdole/gofeed"
)

var MarkdownImageRE = regexp.MustCompile(`!\[\]\(([^)]+)\)`)

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

		if err := processItem(item, downloadDir); err != nil {
			log.Print(err)

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

func processItem(item *gofeed.Item, downloadDir string) error {
	post := Post{} //nolint:exhaustruct

	converter := md.NewConverter("", true, nil)

	body, err := converter.ConvertString(item.Description)
	if err != nil {
		log.Fatal(err)
	}

	post.title = item.Title
	post.body = body

	if err = post.SetDate(item.Published); err != nil {
		return err
	}

	post.FindAttachments(item)
	post.FetchAttachments(downloadDir)

	if item.Extensions["letterboxd"] != nil {
		post.title, *post.date, err = handleLetterboxdExtensions(item, post.title, *post.date)
		if err != nil {
			return err
		}
	}

	if err = invokeDayOne(post.title, post.body, os.Args[2], os.Args[3:], *post.date, *post.attachmentFiles); err != nil {
		return fmt.Errorf("failed invocation of dayone2: %w", err)
	}

	return nil
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

func invokeDayOne(title, body string, journal string, tags []string, date time.Time, attachments []string) error {
	cmdArgs := []string{"new", "--journal", journal, "--isoDate", date.Format("2006-01-02T15:04:05"), "--tags"}
	cmdArgs = append(cmdArgs, tags...)

	for _, i := range attachments {
		cmdArgs = append(cmdArgs, "-a", i)

		defer os.Remove(i)
	}

	post := ""
	if len(title) > 0 {
		post += "# " + title + "\n"
	}

	if len(attachments) > 0 {
		post += "[{attachment}]\n"
	}

	post += body

	cmd := exec.Command("dayone2", cmdArgs...)
	cmd.Stdin = strings.NewReader(post)

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
