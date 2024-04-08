package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setup() (*config, error) {
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.TimeOnly
	}))

	cfg, err := parseConfig()
	if err != nil {
		return nil, err
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("log.ParseLevel: %w", err)
	}

	zerolog.SetGlobalLevel(level)

	return cfg, nil
}

type config struct {
	Port          int               `env:"PORT" envDefault:"4321"`
	Address       string            `env:"ADDRESS,expand" envDefault:":$PORT"`
	LogLevel      string            `env:"LOG_LEVEL" envDefault:"debug"`
	ResponsesDir  string            `env:"RESPONSES_DIR" envDefault:"./.sms_responses"`
	Ext2MIMEType  map[string]string `env:"EXTENSION_MIME_TYPE_MAP"`
	Method2Status map[string]int    `env:"METHOD_STATUS_MAP" envDefault:"DELETE:202,GET:200,PATCH:204,POST:201,PUT:204"`
}

func parseConfig() (*config, error) {
	var cfg config
	return &cfg, env.Parse(&cfg)
}
