package main

import (
	"testing"
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

			if actualOutput := post.Render(); actualOutput != tc.expectedOutput {
				t.Errorf("got %q want %q", actualOutput, tc.expectedOutput)
			}
		})
	}
}
