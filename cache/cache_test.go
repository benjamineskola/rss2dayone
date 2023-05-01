package cache //nolint:testpackage

import (
	"bytes"
	"testing"
)

type TestFile struct {
	data   []byte
	buffer *bytes.Buffer
}

func NewTestFile(data []byte) *TestFile {
	return &TestFile{data: data, buffer: bytes.NewBuffer(data)}
}

func (t *TestFile) Read(p []byte) (int, error) {
	return t.buffer.Read(p) //nolint:wrapcheck
}

func (t *TestFile) Write(p []byte) (int, error) {
	n, err := t.buffer.Write(p)
	t.data = t.buffer.Bytes()

	return n, err //nolint:wrapcheck
}

func (t *TestFile) Seek(offset int64, _ int) (int64, error) {
	return offset, nil
}

func TestCacheLoad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		cacheData       string
		expectToBeEmpty bool
		// expectedOutput map[string]struct{}
	}{
		{name: "empty file", cacheData: "", expectToBeEmpty: true},
		{name: "empty list", cacheData: "[]", expectToBeEmpty: true},
		{name: "file with one string", cacheData: `["foo"]`, expectToBeEmpty: false},
		{name: "file with several strings", cacheData: `["other", "foo"]`, expectToBeEmpty: false},
	}

	for _, tc := range testCases {
		tc := tc //nolint:varnamelen

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cacheBuffer := NewTestFile([]byte(tc.cacheData))

			cache, err := InitWithFile(cacheBuffer)
			if err != nil {
				t.Errorf("failed to initialise cache: %s", err)
			}

			if tc.expectToBeEmpty {
				if len(*cache.ids) > 0 {
					t.Errorf("got %d want 0", len(*cache.ids))
				}
			} else {
				if !cache.Contains("foo") {
					t.Errorf("should contain \"foo\": %q", *cache.ids)
				}
				if cache.Contains("bar") {
					t.Errorf("should not contain \"bar\": %q", *cache.ids)
				}
			}
		})
	}
}

func TestCacheAddSave(t *testing.T) {
	t.Parallel()

	cacheBuffer := NewTestFile([]byte{})

	cache, err := InitWithFile(cacheBuffer)
	if err != nil {
		t.Errorf("failed to initialise cache: %s", err)
	}

	if cache.Contains("foo") {
		t.Errorf("should not contain \"foo\": %q", *cache.ids)
	}

	cache.Add("foo")

	if !cache.Contains("foo") {
		t.Errorf("should contain \"foo\": %q", *cache.ids)
	}

	if string(cacheBuffer.data) != "" {
		t.Errorf("should be empty: %q", cacheBuffer.data)
	}

	err = cache.Save()
	if err != nil {
		t.Errorf("error saving: %s", err)
	}

	if string(cacheBuffer.data) != `["foo"]` {
		t.Errorf("should contain \"foo\": %q", cacheBuffer.data)
	}
}
