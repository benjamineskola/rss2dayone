package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(item.Description)
	if err != nil {
		log.Fatal(err)
	}

	postTime, err := time.Parse("Mon, _2 Jan 2006 15:04:05 -0700", item.Published)
	if err != nil {
		postTime, err = time.Parse(time.RFC3339, item.Published)
		if err != nil {
			log.Fatalf("Could not parse time of %s: %s", item.GUID, err)
		}
	}

	attachmentUrls := findAttachments(item, markdown)

	attachmentFiles := []string{}

	for url := range attachmentUrls {
		file, err := fetchAttachment(url, downloadDir)
		if err != nil {
			log.Print(err)

			continue
		}

		attachmentFiles = append(attachmentFiles, file.Name())
		markdown = strings.ReplaceAll(markdown, "![]("+url+")", "")
	}

	title := item.Title

	if item.Extensions["letterboxd"] != nil {
		title = fmt.Sprintf("%s (%s)",
			item.Extensions["letterboxd"]["filmTitle"][0].Value,
			item.Extensions["letterboxd"]["filmYear"][0].Value)
		postTime, err = time.Parse("2006-01-02", item.Extensions["letterboxd"]["watchedDate"][0].Value)

		if err != nil {
			return fmt.Errorf("could not parse time of %s: %w", item.GUID, err)
		}
	}

	if err = invokeDayOne(title, markdown, os.Args[2], os.Args[3:], postTime, attachmentFiles); err != nil {
		return fmt.Errorf("failed invocation of dayone2: %w", err)
	}

	return nil
}

func findAttachments(item *gofeed.Item, body string) map[string]struct{} {
	attachmentUrls := make(map[string]struct{})

	for _, enclosure := range item.Enclosures {
		attachmentUrls[enclosure.URL] = struct{}{}
	}

	for _, enclosure := range item.Extensions["media"]["content"] {
		attachmentUrls[enclosure.Attrs["url"]] = struct{}{}
	}

	embeddedImages := MarkdownImageRE.FindAllStringSubmatch(body, -1)
	for _, match := range embeddedImages {
		if len(match) > 1 && len(match[1]) > 0 {
			attachmentUrls[match[1]] = struct{}{}
		}
	}

	return attachmentUrls
}

func fetchAttachment(url, downloadDir string) (*os.File, error) {
	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return nil, fmt.Errorf("error downloading attachment: %w", err)
	}

	fileName := strings.ReplaceAll(url, "/", "-")
	fileName = strings.ReplaceAll(fileName, ":", "-")
	fileName, _, _ = strings.Cut(fileName, "?")

	if !(strings.HasSuffix(fileName, ".jpg") ||
		strings.HasSuffix(fileName, ".jpeg") ||
		strings.HasSuffix(fileName, ".png")) {
		fileName += ".jpg" // Not sure that Day One actually cares that this is right, but there has to be one
	}

	defer resp.Body.Close()

	file, err := os.Create(downloadDir + "/" + fileName)
	if err != nil {
		return nil, fmt.Errorf("error downloading attachment: %w", err)
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error saving attachment file: %w", err)
	}

	return file, nil
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
