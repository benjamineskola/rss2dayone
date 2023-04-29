package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var CACHEFILE = filepath.Join(xdg.CacheHome, "rss2dayone.json") //nolint:gochecknoglobals

type Cache struct {
	ids *map[string]struct{}
}

func Init() (*Cache, error) {
	idList := make([]string, 0)

	data, err := os.ReadFile(CACHEFILE)
	if err == nil {
		err = json.Unmarshal(data, &idList)
		if err != nil {
			return nil, fmt.Errorf("error decoding processed map: %w", err)
		}
	}

	ids := make(map[string]struct{})
	for _, i := range idList {
		ids[i] = struct{}{}
	}

	return &Cache{ids: &ids}, nil
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

	if err = os.WriteFile(CACHEFILE, data, 0o600); err != nil {
		return fmt.Errorf("error writing seen data to file: %w", err)
	}

	return nil
}
