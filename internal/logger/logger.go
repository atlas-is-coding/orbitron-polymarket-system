package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// NewWithWriter creates a zerolog logger that writes to the provided writer.
func NewWithWriter(level, format string, w io.Writer) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	var out io.Writer
	if format == "pretty" {
		out = zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	} else {
		out = w
	}
	return zerolog.New(out).Level(lvl).With().Timestamp().Logger()
}

// New создаёт настроенный zerolog логгер.
// format: "pretty" — цветной вывод в stdout (для разработки)
//
//	"json"   — JSON-строки (для продакшена/loki/elk)
func New(level, format string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	var w io.Writer
	if format == "pretty" {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		w = os.Stdout
	}

	return zerolog.New(w).
		Level(lvl).
		With().
		Timestamp().
		Logger()
}
