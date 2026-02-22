package engine

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/nicholas/glint/pkg/rule"
)

type Cache struct {
	mu      sync.RWMutex
	dir     string
	enabled bool
}

type cachedResult struct {
	FileHash    string
	Diagnostics []rule.Diagnostic
}

func NewCache(dir string, enabled bool) (*Cache, error) {
	if enabled && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating cache dir: %w", err)
		}
	}
	return &Cache{dir: dir, enabled: enabled}, nil
}

func HashFile(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (c *Cache) cacheKey(filePath, ruleSet string) string {
	h := sha256.Sum256([]byte(filePath + "\x00" + ruleSet))
	return hex.EncodeToString(h[:16])
}

func (c *Cache) Lookup(filePath, fileHash, ruleSet string) ([]rule.Diagnostic, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(filePath, ruleSet)
	path := filepath.Join(c.dir, key+".gob")

	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	var cr cachedResult
	if err := gob.NewDecoder(f).Decode(&cr); err != nil {
		return nil, false
	}

	if cr.FileHash != fileHash {
		return nil, false
	}

	return cr.Diagnostics, true
}

func (c *Cache) Store(filePath, fileHash, ruleSet string, diags []rule.Diagnostic) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.cacheKey(filePath, ruleSet)
	path := filepath.Join(c.dir, key+".gob")

	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	_ = gob.NewEncoder(f).Encode(cachedResult{
		FileHash:    fileHash,
		Diagnostics: diags,
	})
}

func (c *Cache) Clear() error {
	if c.dir == "" {
		return nil
	}
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".gob" {
			os.Remove(filepath.Join(c.dir, e.Name()))
		}
	}
	return nil
}
