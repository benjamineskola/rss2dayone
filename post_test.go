package main

import (
	"testing"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		title          string
		body           string
		attachments    *[]string
		expectedOutput string
	}{
		{
			name:           "simple title-only post",
			title:          "Hello World",
			expectedOutput: "# Hello World\n",
		},
		{
			name:           "post with title and body",
			title:          "Hello World",
			body:           "Some content beep boop",
			expectedOutput: "# Hello World\nSome content beep boop",
		},
		{
			name:           "post with attachment",
			title:          "Hello World",
			body:           "Some content beep boop",
			expectedOutput: "# Hello World\n[{attachment}]\nSome content beep boop",
			attachments:    &[]string{"hello world"},
		},
	}

	for _, tc := range testCases {
		tc := tc //nolint:varnamelen

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			post := Post{ //nolint:exhaustruct
				title:           tc.title,
				body:            tc.body,
				AttachmentFiles: &[]string{},
			}

			if tc.attachments != nil {
				post.AttachmentFiles = tc.attachments
			}

			assert.Equal(t, tc.expectedOutput, post.Render())
		})
	}
}

func TestFindAttachments(t *testing.T) { //nolint:funlen
	t.Parallel()

	testCases := []struct {
		name           string
		post           *Post
		expectedOutput *map[string]struct{}
	}{
		{
			name:           "no attachments",
			post:           &Post{feedItem: &gofeed.Item{}}, //nolint:exhaustruct
			expectedOutput: &map[string]struct{}{},
		},
		{
			name: "with enclosure",
			post: &Post{feedItem: &gofeed.Item{ //nolint:exhaustruct
				Enclosures: []*gofeed.Enclosure{{URL: "http://test.invalid/hello.jpg"}},
			}},
			expectedOutput: &map[string]struct{}{"http://test.invalid/hello.jpg": {}},
		},
		{
			name: "with media extension",
			post: &Post{feedItem: &gofeed.Item{ //nolint:exhaustruct
				Extensions: ext.Extensions{"media": {
					"content": {
						ext.Extension{Attrs: map[string]string{"url": "http://test.invalid/hello.jpg"}}, //nolint:exhaustruct
					},
				}},
			}},
			expectedOutput: &map[string]struct{}{"http://test.invalid/hello.jpg": {}},
		},
		{
			name:           "with body link",
			post:           &Post{feedItem: &gofeed.Item{}, body: "![](http://test.invalid/hello.jpg)"}, //nolint:exhaustruct
			expectedOutput: &map[string]struct{}{"http://test.invalid/hello.jpg": {}},
		},
		{
			name: "with body link and enclosure",
			post: &Post{ //nolint:exhaustruct
				feedItem: &gofeed.Item{ //nolint:exhaustruct
					Enclosures: []*gofeed.Enclosure{{URL: "http://test.invalid/hello.jpg"}},
				},
				body: "![](http://test.invalid/anotherpicture.jpg)",
			},
			expectedOutput: &map[string]struct{}{
				"http://test.invalid/hello.jpg":          {},
				"http://test.invalid/anotherpicture.jpg": {},
			},
		},
		{
			name: "with body link duplicating an enclosure",
			post: &Post{ //nolint:exhaustruct
				feedItem: &gofeed.Item{ //nolint:exhaustruct
					Enclosures: []*gofeed.Enclosure{{URL: "http://test.invalid/hello.jpg"}},
				},
				body: "![](http://test.invalid/hello.jpg)",
			},
			expectedOutput: &map[string]struct{}{
				"http://test.invalid/hello.jpg": {},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc //nolint:varnamelen

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.post.FindAttachments()

			assert.Equal(t, tc.expectedOutput, tc.post.attachmentUrls)
		})
	}
}
