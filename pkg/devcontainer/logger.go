package devcontainer

import (
	"bufio"
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	Log(src io.Reader) error
}

type LoggerImpl struct {
	logger zerolog.Logger
	level  zerolog.Level
}

func NewLogger(logger zerolog.Logger, level zerolog.Level) Logger {
	return &LoggerImpl{logger: logger, level: level}
}

func (l *LoggerImpl) Log(src io.Reader) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		log.Info().Msg(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func NopLogger() Logger {
	return &LoggerImpl{logger: zerolog.Nop(), level: zerolog.TraceLevel}
}
