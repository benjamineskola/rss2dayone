package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Cache struct {
	ids  *map[string]struct{}
	file string
}

func Init() (*Cache, error) {
	idList := make([]string, 0)
	idSet := make(map[string]struct{})

	cache := Cache{
		ids:  &idSet,
		file: filepath.Join(xdg.CacheHome, "rss2dayone.json"),
	}

	data, err := os.ReadFile(cache.file)
	if err == nil {
		err = json.Unmarshal(data, &idList)
		if err != nil {
			return nil, fmt.Errorf("error decoding processed map: %w", err)
		}
	}

	for _, i := range idList {
		(*cache.ids)[i] = struct{}{}
	}

	return &cache, nil
}

func (cache *Cache) Add(id string) {
	(*cache.ids)[id] = struct{}{}
}

func (cache *Cache) Contains(id string) bool {
	_, isPresent := (*cache.ids)[id]

	return isPresent
}

func (cache *Cache) Save() error {
	idList := make([]string, 0)
	for id := range *cache.ids {
		idList = append(idList, id)
	}

	data, err := json.MarshalIndent(idList, "", "  ")
	if err != nil {
		return fmt.Errorf("error serialising seen data: %w", err)
	}

	if err = os.WriteFile(cache.file, data, 0o600); err != nil {
		return fmt.Errorf("error writing seen data to file: %w", err)
	}

	return nil
}
