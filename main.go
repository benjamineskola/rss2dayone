package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/adrg/xdg"
	"github.com/mmcdole/gofeed"
)

var CACHEFILE = filepath.Join(xdg.CacheHome, "rss2dayone.json") //nolint:gochecknoglobals

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

	processed := loadProcessedItemsList()
	for _, item := range feed.Items {
		_, isPresent := processed[item.GUID]
		if isPresent {
			continue
		}

		if err := processItem(item, downloadDir); err != nil {
			log.Print(err)

			continue
		}

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

	attachmentUrls := []string{}
	for _, enclosure := range item.Enclosures {
		attachmentUrls = append(attachmentUrls, enclosure.URL)
	}

	for _, enclosure := range item.Extensions["media"]["content"] {
		attachmentUrls = append(attachmentUrls, enclosure.Attrs["url"])
	}

	attachmentFiles := []string{}

	for _, url := range attachmentUrls {
		file, err := fetchAttachment(url, downloadDir)
		if err != nil {
			log.Print(err)

			continue
		}

		attachmentFiles = append(attachmentFiles, file.Name())
		markdown = strings.ReplaceAll(markdown, "![]("+url+")", "")
	}

	if err = invokeDayOne(item.Title, markdown, os.Args[2], os.Args[3:], postTime, attachmentFiles); err != nil {
		return fmt.Errorf("failed invocation of dayone2: %w", err)
	}

	return nil
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

	if len(attachments) > 0 {
		body = "[{attachment}]\n" + body
	}

	cmd := exec.Command("dayone2", cmdArgs...)
	cmd.Stdin = strings.NewReader("# " + title + "\n" + body)

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

func saveProcessedItemsList(processed *map[string]struct{}) {
	processedList := make([]string, 0)
	for i := range *processed {
		processedList = append(processedList, i)
	}

	data, err := json.MarshalIndent(processedList, "", "  ")
	if err != nil {
		log.Fatal("Error serialising seen data:", err)
	}

	if err = os.WriteFile(CACHEFILE, data, 0o600); err != nil {
		log.Fatal("Error writing seen data to file:", err)
	}
}
