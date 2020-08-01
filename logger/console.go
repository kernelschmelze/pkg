package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh/terminal"
)

/*
Black        0;30     Dark Gray     1;30
Red          0;31     Light Red     1;31
Green        0;32     Light Green   1;32
Brown/Orange 0;33     Yellow        1;33
Blue         0;34     Light Blue    1;34
Purple       0;35     Light Purple  1;35
Cyan         0;36     Light Cyan    1;36
Light Gray   0;37     White         1;37
*/

const (
	red     = "\033[0;31;1m"
	green   = "\033[0;32m"
	yellow  = "\033[0;33m"
	blue    = "\033[0;34;1m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"
	white   = "\033[0;37m"

	darkGray = "\033[1;30m"

	end = "\033[0m"
)

var (
	noColor    = true
	pathPrefix = ""
)

func init() {
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		noColor = false
	}
	pathPrefix, _ = os.Getwd()
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}
}

func colorize(s interface{}, c string, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%v", s)
	}
	return fmt.Sprintf("%s%v%s", c, s, end)
}

func getConsoleWriter() zerolog.ConsoleWriter {

	output := zerolog.ConsoleWriter{
		TimeFormat: timeFormat,
		NoColor:    noColor,
		Out:        os.Stderr,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
			zerolog.CallerFieldName,
		},
	}

	output.FormatCaller = func(i interface{}) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
			if offset := strings.Index(c, pathPrefix); offset >= 0 {
				c = cc[offset+len(pathPrefix):]
			}
			c = colorize(c, darkGray, noColor)
		}
		return c
	}

	output.FormatLevel = func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case "debug":
				l = colorize("DBG", blue, noColor)
			case "info":
				l = colorize("INF", green, noColor)
			case "warn":
				l = colorize("WRN", yellow, noColor)
			case "error":
				l = colorize("ERR", red, noColor)
			case "fatal":
				l = colorize("FTL", red, noColor)
			case "panic":
				l = colorize("???", magenta, noColor)
			default:
				l = colorize("???", magenta, noColor)
			}
		} else {
			l = strings.ToUpper(fmt.Sprintf("%s", i))
			if len(l) > 3 {
				l = l[0:3]
			}
		}
		return l
	}

	output.FormatFieldName = func(i interface{}) string {
		return colorize(fmt.Sprintf("%s=", i), cyan, noColor)
	}

	output.FormatFieldValue = func(i interface{}) string {
		return colorize(fmt.Sprintf("%s", i), blue, noColor)
	}

	return output
}
