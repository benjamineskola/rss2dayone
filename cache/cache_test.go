package cache //nolint:testpackage

import (
	"bytes"
	// "os".
	"testing"

	"github.com/stretchr/testify/assert"
)

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

			cacheBuffer := bytes.NewBuffer([]byte(tc.cacheData))

			cache, err := InitWithBuffer(cacheBuffer, "")
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedOutput, *cache.ids)
			}
		})
	}
}

func TestCacheAddSave(t *testing.T) {
	t.Parallel()

	cacheBuffer := bytes.NewBuffer([]byte{})

	cache, err := InitWithBuffer(cacheBuffer, "")

	assert.NoError(t, err)

	assert.False(t, cache.Contains("foo"))

	cache.Add("foo")

	assert.True(t, cache.Contains("foo"))

	assert.Equal(t, []byte{}, cacheBuffer.Bytes())

	assert.NoError(t, cache.SaveToBuffer(cacheBuffer))

	assert.Equal(t, []byte(`["foo"]`), cacheBuffer.Bytes())
}
