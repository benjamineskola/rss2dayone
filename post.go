package main

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

type Post struct {
	title string
	body  string
	date  *time.Time

	AttachmentUrls *map[string]struct{}
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
	p.AttachmentUrls = &m

	for _, enclosure := range item.Enclosures {
		(*p.AttachmentUrls)[enclosure.URL] = struct{}{}
	}

	for _, enclosure := range item.Extensions["media"]["content"] {
		(*p.AttachmentUrls)[enclosure.Attrs["url"]] = struct{}{}
	}

	embeddedImages := MarkdownImageRE.FindAllStringSubmatch(p.body, -1)
	for _, match := range embeddedImages {
		if len(match) > 1 && len(match[1]) > 0 {
			(*p.AttachmentUrls)[match[1]] = struct{}{}
		}
	}
}
