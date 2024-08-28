package devcontainer

import (
	"io"

	"github.com/rs/zerolog"
)

type Logger interface {
	Log(src io.Reader)
}

type LoggerImpl struct {
	logger zerolog.Logger
	level  zerolog.Level
}

func NewLogger(logger zerolog.Logger, level zerolog.Level) Logger {
	return &LoggerImpl{logger: logger, level: level}
}

func (l *LoggerImpl) Log(src io.Reader) {
	content, err := io.ReadAll(src)
	if err != nil {
		l.logger.Error().Err(err).Msg("Failed to log content")
	}

	l.logger.WithLevel(l.level).Msg(string(content))
}

func NopLogger() Logger {
	return &LoggerImpl{logger: zerolog.Nop(), level: zerolog.TraceLevel}
}
