package wzerolog

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90
)

func newConsoleWriter() zerolog.ConsoleWriter {
	console := zerolog.ConsoleWriter{Out: os.Stdout}
	console.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("[ %s ]", i)
	}
	console.FormatFieldName = func(i interface{}) string {
		return colorize(colorize(fmt.Sprintf("%s:", i), colorCyan), colorBold)
	}
	console.FormatTimestamp = consoleFormatTimestamp()
	return console
}

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func colorize(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func consoleFormatTimestamp() func(i interface{}) string {
	timeFormat := "2006/01/02 15:04:05"
	return func(i interface{}) string {
		t := "<nil>"
		switch tt := i.(type) {
		case string:
			ts, err := time.Parse(zerolog.TimeFieldFormat, tt)
			if err != nil {
				t = tt
			} else {
				t = ts.Format(timeFormat)
			}
		case json.Number:
			i, err := tt.Int64()
			if err != nil {
				t = tt.String()
			} else {
				var sec, nsec int64 = i, 0
				switch zerolog.TimeFieldFormat {
				case zerolog.TimeFormatUnixMs:
					nsec = int64(time.Duration(i) * time.Millisecond)
					sec = 0
				case zerolog.TimeFormatUnixMicro:
					nsec = int64(time.Duration(i) * time.Microsecond)
					sec = 0
				}
				ts := time.Unix(sec, nsec).Local()
				t = ts.Format(timeFormat)
			}
		}
		return colorize(t, colorBlue)
	}
}
