package engine

import (
	"testing"

	"github.com/mjc-gh/virgo"
	"github.com/mjc-gh/virgo/internal/pagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformScreenshotTask(t *testing.T) {
	server := pagetest.NewTestWebServer("simple")
	task := NewTask("screenshot", server.URL)

	ctx, cancel := pagetest.NewTestContext()
	defer cancel()

	sr, err := performScreenshotTask(ctx, &task, virgo.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, sr.Buffer)
}
