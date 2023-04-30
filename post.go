package main

import (
	"fmt"
	"time"
)

type Post struct {
	title string
	body  string
	date  *time.Time
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
