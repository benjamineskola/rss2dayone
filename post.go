package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
)

type Post struct {
	title string
	body  string
	date  *time.Time

	attachmentUrls  *map[string]struct{}
	AttachmentFiles *[]string
}

func NewPost(item *gofeed.Item, downloadDir string) (*Post, error) {
	post := Post{ //nolint:exhaustruct
		title: item.Title,
	}
	converter := md.NewConverter("", true, nil)

	body, err := converter.ConvertString(item.Description)
	if err != nil {
		log.Fatal(err)
	}

	post.body = body

	if err = post.SetDate(item.Published); err != nil {
		return nil, err
	}

	post.FindAttachments(item)
	post.FetchAttachments(downloadDir)

	if item.Extensions["letterboxd"] != nil {
		post.title, *post.date, err = handleLetterboxdExtensions(item, post.title, *post.date)
		if err != nil {
			return nil, err
		}
	}

	return &post, nil
}

func (p *Post) SetDate(date string) error {
	res, err := time.Parse("Mon, _2 Jan 2006 15:04:05 -0700", date)
	if err != nil {
		res, err = time.Parse(time.RFC3339, date)
		if err != nil {
			return fmt.Errorf("could not parse date %q: %w", date, err)
		}
	}

	p.date = &res

	return nil
}

func (p *Post) FindAttachments(item *gofeed.Item) {
	m := make(map[string]struct{})
	p.attachmentUrls = &m

	for _, enclosure := range item.Enclosures {
		(*p.attachmentUrls)[enclosure.URL] = struct{}{}
	}

	for _, enclosure := range item.Extensions["media"]["content"] {
		(*p.attachmentUrls)[enclosure.Attrs["url"]] = struct{}{}
	}

	embeddedImages := MarkdownImageRE.FindAllStringSubmatch(p.body, -1)
	for _, match := range embeddedImages {
		if len(match) > 1 && len(match[1]) > 0 {
			(*p.attachmentUrls)[match[1]] = struct{}{}
		}
	}
}

func (p *Post) FetchAttachments(downloadDir string) {
	a := []string{}
	p.AttachmentFiles = &a

	for url := range *p.attachmentUrls {
		file, err := fetchAttachment(url, downloadDir)
		if err != nil {
			log.Print(err)

			continue
		}

		*p.AttachmentFiles = append(*p.AttachmentFiles, file.Name())
		p.body = strings.ReplaceAll(p.body, "![]("+url+")", "")
	}
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
