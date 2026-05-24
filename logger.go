package virgo

import (
	"io"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog"
)

var logger *zerolog.Logger

func SetupLogger(debug bool, jsonLogs bool) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	var writer io.Writer
	if jsonLogs {
		// JSON logs output to stdout
		writer = os.Stdout
	} else {
		// Create a multi-level writer that routes logs to stdout/stderr based on level
		writer = zerolog.MultiLevelWriter(
			SpecificLevelWriter{
				Writer: zerolog.ConsoleWriter{Out: os.Stdout},
				Levels: []zerolog.Level{
					zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
				},
			},
			SpecificLevelWriter{
				Writer: zerolog.ConsoleWriter{Out: os.Stderr},
				Levels: []zerolog.Level{
					zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
				},
			},
		)
	}

	l := buildLogger(writer)
	logger = &l

	return logger
}

// buildLogger creates a logger with common configuration.
func buildLogger(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().
		Str("source", "go").
		Str("service", "virgo").
		Logger().
		Sample(
			zerolog.LevelSampler{
				TraceSampler: &zerolog.BurstSampler{
					Burst:       1,
					Period:      2 * time.Second,
					NextSampler: &zerolog.BasicSampler{N: 100},
				},
				WarnSampler: &zerolog.BurstSampler{
					Burst:       4,
					Period:      1 * time.Second,
					NextSampler: &zerolog.BasicSampler{N: 100},
				},
			},
		)
}

// SpecificLevelWriter is a LevelWriter that only writes logs for specific levels.
type SpecificLevelWriter struct {
	Writer zerolog.ConsoleWriter
	Levels []zerolog.Level
}

// Write implements io.Writer and writes using the console writer.
func (w SpecificLevelWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

// WriteLevel writes the log message if the level matches one of the configured levels.
func (w SpecificLevelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	if slices.Contains(w.Levels, level) {
		return w.Writer.Write(p)
	}

	return len(p), nil
}

func Logger() *zerolog.Logger {
	if logger == nil {
		nop := zerolog.Nop()

		return &nop
	}

	return logger
}
