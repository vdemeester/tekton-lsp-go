package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

func TestDocumentCache_InsertAndGet(t *testing.T) {
	c := New()

	c.Insert("file:///test.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")

	doc, ok := c.Get("file:///test.yaml")
	require.True(t, ok, "document should be found")
	assert.Equal(t, "file:///test.yaml", doc.URI)
	assert.Equal(t, "yaml", doc.LanguageID)
	assert.Equal(t, int32(1), doc.Version)
	assert.Equal(t, "apiVersion: tekton.dev/v1\nkind: Task\n", doc.Content)
}

func TestDocumentCache_GetParsed(t *testing.T) {
	c := New()

	c.Insert("file:///test.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")

	parsed, ok := c.GetParsed("file:///test.yaml")
	require.True(t, ok, "parsed document should be available")
	assert.Equal(t, "tekton.dev/v1", parsed.APIVersion)
	assert.Equal(t, "Task", parsed.Kind)
}

func TestDocumentCache_GetMissing(t *testing.T) {
	c := New()

	_, ok := c.Get("file:///missing.yaml")
	assert.False(t, ok, "missing document should not be found")

	_, ok = c.GetParsed("file:///missing.yaml")
	assert.False(t, ok, "missing parsed doc should not be found")
}

func TestDocumentCache_Update(t *testing.T) {
	c := New()

	c.Insert("file:///test.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")
	c.Update("file:///test.yaml", 2, "apiVersion: tekton.dev/v1\nkind: Pipeline\n")

	doc, ok := c.Get("file:///test.yaml")
	require.True(t, ok)
	assert.Equal(t, int32(2), doc.Version)
	assert.Equal(t, "apiVersion: tekton.dev/v1\nkind: Pipeline\n", doc.Content)

	parsed, ok := c.GetParsed("file:///test.yaml")
	require.True(t, ok)
	assert.Equal(t, "Pipeline", parsed.Kind, "parsed cache should be updated after content change")
}

func TestDocumentCache_Remove(t *testing.T) {
	c := New()

	c.Insert("file:///test.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")
	require.True(t, func() bool { _, ok := c.Get("file:///test.yaml"); return ok }())

	c.Remove("file:///test.yaml")

	_, ok := c.Get("file:///test.yaml")
	assert.False(t, ok, "removed document should not be found")

	_, ok = c.GetParsed("file:///test.yaml")
	assert.False(t, ok, "removed parsed doc should not be found")
}

func TestDocumentCache_All(t *testing.T) {
	c := New()

	c.Insert("file:///a.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")
	c.Insert("file:///b.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Pipeline\n")

	all := c.All()
	assert.Len(t, all, 2, "should have 2 documents")
}

func TestDocumentCache_AllParsed(t *testing.T) {
	c := New()

	c.Insert("file:///a.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")
	c.Insert("file:///b.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Pipeline\n")

	all := c.AllParsed()
	assert.Len(t, all, 2, "should have 2 parsed documents")

	kinds := make(map[string]bool)
	for _, doc := range all {
		kinds[doc.Kind] = true
	}
	assert.True(t, kinds["Task"], "should contain Task")
	assert.True(t, kinds["Pipeline"], "should contain Pipeline")
}

func TestDocumentCache_ConcurrentAccess(t *testing.T) {
	c := New()

	// Simulate concurrent reads/writes
	done := make(chan bool, 10)
	for i := range 5 {
		go func(i int) {
			uri := "file:///test.yaml"
			c.Insert(uri, "yaml", int32(i), "apiVersion: tekton.dev/v1\nkind: Task\n")
			c.Get(uri)
			c.GetParsed(uri)
			done <- true
		}(i)
	}
	for range 5 {
		<-done
	}

	// Should not panic or deadlock
	_, ok := c.Get("file:///test.yaml")
	assert.True(t, ok)
}

// Ensure the cache uses our parser types correctly.
func TestDocumentCache_ParserIntegration(t *testing.T) {
	// Directly parse to verify compatibility
	doc, err := parser.ParseYAML("test.yaml", "apiVersion: tekton.dev/v1\nkind: Task\n")
	require.NoError(t, err)
	assert.Equal(t, "Task", doc.Kind)

	// Cache should give same results
	c := New()
	c.Insert("file:///test.yaml", "yaml", 1, "apiVersion: tekton.dev/v1\nkind: Task\n")

	parsed, ok := c.GetParsed("file:///test.yaml")
	require.True(t, ok)
	assert.Equal(t, doc.Kind, parsed.Kind)
	assert.Equal(t, doc.APIVersion, parsed.APIVersion)
}
