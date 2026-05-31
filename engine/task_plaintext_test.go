package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mjc-gh/virgo"
	"github.com/mjc-gh/virgo/internal/pagetest"
)

func TestPerformPlaintextTask(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// With default selector ("body"), should contain all body content
	assert.Contains(t, pr.Content, "Main Content Heading")
	assert.Contains(t, pr.Content, "Article Content")
	assert.Contains(t, pr.Content, "Footer content here")
	assert.Contains(t, pr.Content, "Extra content in body")
}

func TestPerformPlaintextTaskDefaultSelector(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL, WithParams(map[string]any{}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// Default selector should be "body", containing all content
	assert.Contains(t, pr.Content, "Main Content Heading")
	assert.Contains(t, pr.Content, "Article Content")
}

func TestPerformPlaintextTaskCustomSelectorMain(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL, WithParams(map[string]any{
		"selector": "main",
	}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// Should contain only main content
	assert.Contains(t, pr.Content, "Main Content Heading")
	assert.Contains(t, pr.Content, "This is the main content here")
	assert.Contains(t, pr.Content, "Additional paragraph in main")

	// Should NOT contain content outside main
	assert.NotContains(t, pr.Content, "Footer content")
	assert.NotContains(t, pr.Content, "Article Content")
}

func TestPerformPlaintextTaskCustomSelectorByID(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL, WithParams(map[string]any{
		"selector": "#content",
	}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// Should contain only article content (by ID)
	assert.Contains(t, pr.Content, "Article Content")
	assert.Contains(t, pr.Content, "This is article content for testing ID selectors")

	// Should NOT contain other content
	assert.NotContains(t, pr.Content, "Main Content Heading")
	assert.NotContains(t, pr.Content, "Footer content")
}

func TestPerformPlaintextTaskCustomSelectorByClass(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL, WithParams(map[string]any{
		"selector": ".article",
	}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// Should contain only article content (by class)
	assert.Contains(t, pr.Content, "Article Content")
	assert.Contains(t, pr.Content, "This is article content for testing ID selectors")
}

func TestPerformPlaintextTaskSelectorFooter(t *testing.T) {
	t.Parallel()
	server := pagetest.NewTestWebServer("plaintext")
	task := NewTask("plaintext", server.URL, WithParams(map[string]any{
		"selector": "footer",
	}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	pr, err := performPlaintextTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, pr.Content)

	// Should contain only footer content
	assert.Contains(t, pr.Content, "Footer content here")

	// Should NOT contain other content
	assert.NotContains(t, pr.Content, "Main Content Heading")
	assert.NotContains(t, pr.Content, "Article Content")
}
