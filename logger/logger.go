package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var (
	timeFormat = "0102 15:04:05.000000"
	logger     = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
)

type Logger struct {
}

func newLogger() *Logger {
	zerolog.CallerSkipFrameCount = 4
	return &Logger{}
}

func init() {

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = timeFormat

	output := getConsoleWriter()
	logger = logger.Output(output)

}

func (l *Logger) Error(v ...interface{}) {
	logger.Error().Msgf("%s", v...)
}

func (l *Logger) Errorf(f string, v ...interface{}) {
	logger.Error().Msgf(f, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	logger.Warn().Msgf("%s", v...)
}

func (l *Logger) Warnf(f string, v ...interface{}) {
	logger.Warn().Msgf(f, v...)
}

func (l *Logger) Info(v ...interface{}) {
	logger.Info().Msgf("%s", v...)
}

func (l *Logger) Infof(f string, v ...interface{}) {
	logger.Info().Msgf(f, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	logger.Debug().Msgf("%s", v...)
}

func (l *Logger) Debugf(f string, v ...interface{}) {
	logger.Debug().Msgf(f, v...)
}
