package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(item.Description)
	if err != nil {
		log.Fatal(err)
	}

	time, err := time.Parse(time.RFC1123Z, item.Published)
	if err != nil {
		log.Fatalf("Could not parse time of %s: %s", item.GUID, err)
	}

	if err = invokeDayOne(markdown, os.Args[2], os.Args[3:], time); err != nil {
		log.Fatalf("Failed invocation of dayone2: %s", err)
	}
}

func invokeDayOne(body string, journal string, tags []string, date time.Time) error {
	cmdArgs := []string{"new", "--journal", journal, "--isoDate", date.Format("2006-01-02T15:04:05"), "--tags"}
	cmdArgs = append(cmdArgs, tags...)
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
