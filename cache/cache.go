package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Cache struct {
	ids      *map[string]struct{}
	path     string
	file     io.ReadWriteSeeker
	modified bool
}

func Init() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cannot identify user cache dir: %w", err)
	}

	path := filepath.Join(cacheDir, "rss2dayone.json")
	file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)

	return InitWithFile(file, path)
}

func InitWithFile(file io.ReadWriteSeeker, path string) (*Cache, error) {
	idList := make([]string, 0)
	idSet := make(map[string]struct{})

	cache := Cache{
		ids:      &idSet,
		path:     path,
		file:     file,
		modified: false,
	}

	var data bytes.Buffer
	if _, err := io.Copy(&data, cache.file); err != nil {
		return nil, fmt.Errorf("error loading data: %w", err)
	}

	if data.Len() == 0 {
		return &cache, nil
	}

	if err := json.Unmarshal(data.Bytes(), &idList); err != nil {
		return nil, fmt.Errorf("error decoding processed map: %w", err)
	}

	for _, i := range idList {
		(*cache.ids)[i] = struct{}{}
	}

	return &cache, nil
}

func (cache *Cache) Add(id string) {
	if !cache.Contains(id) {
		cache.modified = true
	}

	(*cache.ids)[id] = struct{}{}
}

func (cache *Cache) Contains(id string) bool {
	_, isPresent := (*cache.ids)[id]

	return isPresent
}

func (cache *Cache) Save() error {
	if !cache.modified {
		return nil
	}

	idList := make([]string, 0)
	for id := range *cache.ids {
		idList = append(idList, id)
	}

	data, err := json.Marshal(idList)
	if err != nil {
		return fmt.Errorf("error serialising seen data: %w", err)
	}

	_, err = cache.file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error writing seen data: %w", err)
	}

	if cache.path != "" {
		file, err := os.OpenFile(cache.path, os.O_TRUNC, 0)
		if err != nil {
			return fmt.Errorf("error truncating cache file: %w", err)
		}

		_ = file.Truncate(int64(len(data)))
		file.Close()
	}

	_, err = cache.file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing seen data: %w", err)
	}

	return nil
}
