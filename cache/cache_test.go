package cache //nolint:testpackage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
		name           string
		cacheData      string
		expectedOutput map[string]struct{}
	}{
		{name: "empty file", cacheData: "", expectedOutput: map[string]struct{}{}},
		{name: "empty list", cacheData: "[]", expectedOutput: map[string]struct{}{}},
		{name: "file with one string", cacheData: `["foo"]`, expectedOutput: map[string]struct{}{"foo": {}}},
		{
			name: "file with several strings", cacheData: `["other", "foo"]`,
			expectedOutput: map[string]struct{}{"foo": {}, "other": {}},
		},
	}

	for _, tc := range testCases {
		tc := tc //nolint:varnamelen

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cacheBuffer := NewTestFile([]byte(tc.cacheData))

			cache, err := InitWithFile(cacheBuffer)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedOutput, *cache.ids)
			}
		})
	}
}

func TestCacheAddSave(t *testing.T) {
	t.Parallel()

	cacheBuffer := NewTestFile([]byte{})

	cache, err := InitWithFile(cacheBuffer)

	assert.NoError(t, err)

	assert.False(t, cache.Contains("foo"))

	cache.Add("foo")

	assert.True(t, cache.Contains("foo"))

	assert.Equal(t, []byte{}, cacheBuffer.data)

	assert.NoError(t, cache.Save())

	assert.Equal(t, []byte(`["foo"]`), cacheBuffer.data)
}
