package main

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/stretchr/testify/assert"
)

func TestLetterboxdExtensions(t *testing.T) { //nolint:funlen
	t.Parallel()

	now := time.Now().UTC().Round(time.Second)

	testCases := []struct {
		name      string
		input     *gofeed.Item
		wantTitle string
		wantDate  time.Time
	}{
		{
			name: "basic case",
			input: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{
					"letterboxd": {
						"filmTitle":   {ext.Extension{Value: "Film Name"}},  //nolint:exhaustruct
						"filmYear":    {ext.Extension{Value: "2000"}},       //nolint:exhaustruct
						"watchedDate": {ext.Extension{Value: "2023-04-30"}}, //nolint:exhaustruct
					},
				},
			},
			wantTitle: "Film Name (2000)",
			wantDate:  time.Date(2023, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "without title",
			input: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{
					"letterboxd": {
						"filmYear":    {ext.Extension{Value: "2000"}},       //nolint:exhaustruct
						"watchedDate": {ext.Extension{Value: "2023-04-30"}}, //nolint:exhaustruct
					},
				},
			},
			wantTitle: "Original Title",
			wantDate:  time.Date(2023, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "without film year",
			input: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{
					"letterboxd": {
						"filmTitle":   {ext.Extension{Value: "Film Name"}},  //nolint:exhaustruct
						"watchedDate": {ext.Extension{Value: "2023-04-30"}}, //nolint:exhaustruct
					},
				},
			},
			wantTitle: "Original Title",
			wantDate:  time.Date(2023, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "without title or film year",
			input: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{
					"letterboxd": {
						"watchedDate": {ext.Extension{Value: "2023-04-30"}}, //nolint:exhaustruct
					},
				},
			},
			wantTitle: "Original Title",
			wantDate:  time.Date(2023, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "without watched date",
			input: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{
					"letterboxd": {
						"filmTitle": {ext.Extension{Value: "Film Name"}}, //nolint:exhaustruct
						"filmYear":  {ext.Extension{Value: "2000"}},      //nolint:exhaustruct
					},
				},
			},
			wantTitle: "Film Name (2000)",
			wantDate:  now,
		},
	}

	for _, tc := range testCases {
		tc := tc //nolint:varnamelen
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			post := Post{title: "Original Title", feedItem: tc.input} //nolint:exhaustruct
			_ = post.SetDate(now.Format(time.RFC3339))
			_ = post.handleLetterboxdExtensions()

			assert.Equal(t, tc.wantTitle, post.title)
			assert.Equal(t, tc.wantDate, *post.date)
		})
	}
}
