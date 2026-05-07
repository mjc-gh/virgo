package virgo

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logger = nil

	assert.NotNil(t, Logger())
}

func TestSetupLogger(t *testing.T) {
	logger = nil

	l := SetupLogger(false)

	assert.NotNil(t, logger)
	assert.Equal(t, l, Logger())
}

func TestSetupLoggerWithDebug(t *testing.T) {
	_ = SetupLogger(true)

	assert.Equal(t, zerolog.Level(0), zerolog.GlobalLevel())
}
