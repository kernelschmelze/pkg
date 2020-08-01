package logger

var l *Logger

func init() {
	l = newLogger()
}

// DefaultLog provides a default implementation of the Logger interface.
type DefaultLog struct{}

// Logger instances provide custom logging.
type SimpleLogger interface {

	// Log with level ERROR
	Error(...interface{})

	// Log formatted messages with level ERROR
	Errorf(string, ...interface{})

	// Log with level WARN
	Warn(...interface{})

	// Log formatted messages with level WARN
	Warnf(string, ...interface{})

	// Log with level INFO
	Info(...interface{})

	// Log formatted messages with level INFO
	Infof(string, ...interface{})

	// Log with level DEBUG
	Debug(...interface{})

	// Log formatted messages with level DEBUG
	Debugf(string, ...interface{})
}

func (*DefaultLog) Error(a ...interface{})            { l.Error(a...) }
func (*DefaultLog) Errorf(f string, a ...interface{}) { l.Errorf(f, a...) }
func (*DefaultLog) Warn(a ...interface{})             { l.Warn(a...) }
func (*DefaultLog) Warnf(f string, a ...interface{})  { l.Warnf(f, a...) }
func (*DefaultLog) Info(a ...interface{})             { l.Info(a...) }
func (*DefaultLog) Infof(f string, a ...interface{})  { l.Infof(f, a...) }
func (*DefaultLog) Debug(a ...interface{})            { l.Debug(a...) }
func (*DefaultLog) Debugf(f string, a ...interface{}) { l.Debugf(f, a...) }

func Error(a ...interface{})            { l.Error(a...) }
func Errorf(f string, a ...interface{}) { l.Errorf(f, a...) }
func Warn(a ...interface{})             { l.Warn(a...) }
func Warnf(f string, a ...interface{})  { l.Warnf(f, a...) }
func Info(a ...interface{})             { l.Info(a...) }
func Infof(f string, a ...interface{})  { l.Infof(f, a...) }
func Debug(a ...interface{})            { l.Debug(a...) }
func Debugf(f string, a ...interface{}) { l.Debugf(f, a...) }
