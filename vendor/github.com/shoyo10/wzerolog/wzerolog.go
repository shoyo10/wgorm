package wzerolog

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	// LogLevel: 0 debug, 1 info, 2 warn, 3 error
	LogLevel     zerolog.Level `yaml:"log_level" mapstructure:"log_level"`
	PrettyOutput bool          `yaml:"pretty_output" mapstructure:"pretty_output"`
	AppID        string        `yaml:"app_id" mapstructure:"app_id"`
	Env          string        `yaml:"env" mapstructure:"env"`
}

type NoMsgHook struct{}

func (h NoMsgHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if msg == "" {
		e.Str("message", "no message")
	}
}

func Init(cfg Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFieldName = "log_time"
	zerolog.DisableSampling(true)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.SetGlobalLevel(cfg.LogLevel)

	var outWriter io.Writer
	outWriter = os.Stdout

	if cfg.PrettyOutput {
		outWriter = newConsoleWriter()
	}

	logger := zerolog.New(outWriter).
		With().
		Timestamp().
		Caller().
		Str("app_id", cfg.AppID).
		Str("env", cfg.Env).
		Logger().Hook(NoMsgHook{})
	log.Logger = logger
}
