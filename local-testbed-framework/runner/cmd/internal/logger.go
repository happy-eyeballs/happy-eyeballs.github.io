package internal

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	output := io.Writer(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.UnixDate,
	})

	if LogFilePath != "" {
		fileOutput, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(fmt.Sprintf("Cannot open/create log file %q", LogFilePath))
		}

		output = io.MultiWriter(output, fileOutput)
	}

	return zerolog.New(output).
		Level(zerolog.Level(Verbosity)).
		With().
		Timestamp().
		Logger()
}
