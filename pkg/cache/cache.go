package cache

import (
	"sync"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// Entry represents a cached document with its raw content and parsed ASTs.
type Entry struct {
	URI        string
	LanguageID string
	Version    int32
	Content    string
	parsed     []*parser.Document
}

// Cache is a thread-safe cache for open documents and their parsed ASTs.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New creates a new empty document cache.
func New() *Cache {
	return &Cache{
		entries: make(map[string]*Entry),
	}
}

// Insert adds or replaces a document in the cache and parses it.
func (c *Cache) Insert(uri, languageID string, version int32, content string) {
	parsed, _ := parser.ParseAllYAML(uri, content)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[uri] = &Entry{
		URI:        uri,
		LanguageID: languageID,
		Version:    version,
		Content:    content,
		parsed:     parsed,
	}
}

// Get returns the raw document entry for a URI.
func (c *Cache) Get(uri string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[uri]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// GetParsed returns the first parsed document for a URI.
// Deprecated: Use GetAllParsed for multi-document support.
func (c *Cache) GetParsed(uri string) (*parser.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[uri]
	if !ok || len(e.parsed) == 0 {
		return nil, false
	}
	return e.parsed[0], true
}

// GetAllParsed returns all parsed documents for a URI (multi-document YAML).
func (c *Cache) GetAllParsed(uri string) ([]*parser.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[uri]
	if !ok || len(e.parsed) == 0 {
		return nil, false
	}
	return e.parsed, true
}

// Update replaces the content and re-parses the document.
func (c *Cache) Update(uri string, version int32, content string) {
	parsed, _ := parser.ParseAllYAML(uri, content)

	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[uri]; ok {
		e.Version = version
		e.Content = content
		e.parsed = parsed
	}
}

// Remove deletes a document from the cache.
func (c *Cache) Remove(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, uri)
}

// All returns all cached document entries.
func (c *Cache) All() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Entry, 0, len(c.entries))
	for _, e := range c.entries {
		result = append(result, *e)
	}
	return result
}

// AllParsed returns all parsed documents across all cached files.
func (c *Cache) AllParsed() []*parser.Document {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*parser.Document
	for _, e := range c.entries {
		result = append(result, e.parsed...)
	}
	return result
}
