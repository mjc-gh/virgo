package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mjc-gh/virgo"
	"github.com/mjc-gh/virgo/internal/pagetest"
)

func TestPerformMarkdownTask(t *testing.T) {
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	content := mr.Content

	// Test headers are converted
	assert.Contains(t, content, "# Main Heading")
	assert.Contains(t, content, "## Subheading Level 2")
	assert.Contains(t, content, "### Subheading Level 3")
	assert.Contains(t, content, "#### Subheading Level 4")
	assert.Contains(t, content, "##### Subheading Level 5")
	assert.Contains(t, content, "###### Subheading Level 6")

	// Test links are converted
	assert.Contains(t, content, "[link to example](https://example.com)")

	// Test text formatting
	assert.Contains(t, content, "**bold text**")
	assert.Contains(t, content, "*italic text*")
	assert.Contains(t, content, "**bold using b tag**")
	assert.Contains(t, content, "*italic using i tag*")
	assert.Contains(t, content, "<u>underlined text</u>")
	assert.Contains(t, content, "~~strikethrough text~~")

	// Test inline code
	assert.Contains(t, content, "`code example`")

	// Test code blocks
	assert.Contains(t, content, "```")
	assert.Contains(t, content, "function hello()")

	// Test unordered list
	assert.Contains(t, content, "- First item")
	assert.Contains(t, content, "- Second item")
	assert.Contains(t, content, "- Third item")

	// Test ordered list
	assert.Contains(t, content, "1. Step one")
	assert.Contains(t, content, "2. Step two")
	assert.Contains(t, content, "3. Step three")

	// Test blockquote
	assert.Contains(t, content, "> ")
	assert.Contains(t, content, "blockquote")

	// Test image is NOT included by default
	assert.NotContains(t, content, "![Test Image]")

	// Test horizontal rule
	assert.Contains(t, content, "---")

	// Test that hidden elements are excluded
	assert.NotContains(t, content, "display none")
	assert.NotContains(t, content, "visibility hidden")
	assert.NotContains(t, content, "opacity 0")

	// Test that nav/footer elements are excluded
	assert.NotContains(t, content, "Home")
	assert.NotContains(t, content, "About")
	assert.NotContains(t, content, "Footer content")
}

func TestPerformMarkdownTaskSimplePage(t *testing.T) {
	server := pagetest.NewTestWebServer("simple")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	// Simple page has no main/article, should fall back to body
	assert.Contains(t, mr.Content, "Simple Page")
}

func TestPerformMarkdownTaskFallbackToBody(t *testing.T) {
	// Test that when there's no main/article element, body is used
	server := pagetest.NewTestWebServer("simple")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	// The simple page has "Simple Page" in the body, plus "Hello world!" added by script.js
	assert.Contains(t, mr.Content, "Simple Page")
	assert.Contains(t, mr.Content, "Hello world!")
}

func TestPerformMarkdownTaskIncludeImages(t *testing.T) {
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL, WithParams(map[string]any{
		"include-images": true,
	}))

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Test that images ARE included when param is set
	assert.Contains(t, mr.Content, "![Test Image](/images/test.png)")
}

func TestPerformMarkdownTaskNestedLists(t *testing.T) {
	// Test nested lists (ul within ul, ol within ul)
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Test that nested list items are properly formatted
	assert.Contains(t, mr.Content, "- Parent item")
	assert.Contains(t, mr.Content, "- Nested item 1")
	assert.Contains(t, mr.Content, "- Nested item 2")
}

func TestPerformMarkdownTaskAdjacentInlineFormatting(t *testing.T) {
	// Test adjacent inline formatting renders cleanly
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Adjacent bold and italic should render properly
	assert.Contains(t, mr.Content, "**bold***italic*")
}

func TestPerformMarkdownTaskEmptyParagraphCollapse(t *testing.T) {
	// Test that multiple empty paragraphs collapse to single blank line
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Verify that content after empty paragraphs is present
	assert.Contains(t, mr.Content, "Content after empty paragraphs")
	// Verify there are not excessive newlines (3+ newlines should not exist)
	assert.NotContains(t, mr.Content, "\n\n\n")
}

func TestPerformMarkdownTaskBlockquoteWithNestedContent(t *testing.T) {
	// Test complex blockquote with nested formatting and lists
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Blockquote with formatting
	assert.Contains(t, mr.Content, "> ")
	assert.Contains(t, mr.Content, "Quote with")
}

func TestPerformMarkdownTaskHeadersOnNewlines(t *testing.T) {
	// Test that headers are always preceded by a newline (except at start)
	// This tests that inline content followed by a header has proper separation
	server := pagetest.NewTestWebServer("markdown")
	task := NewTask("markdown", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	mr, err := performMarkdownTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, mr.Content)

	// Headers must always be preceded by a newline (except at start)
	// This tests that inline content followed by a header has proper separation
	assert.Contains(t, mr.Content, "text before heading\n\n# Inline to Block Test")

	// Verify no headers appear immediately after inline content (single newline is not enough)
	// Headers should have double newlines before them
	assert.NotRegexp(t, `[^\n]\n#+ `, mr.Content)
}
