package main

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	log "github.com/sirupsen/logrus"
)

func setup() (*config, error) {
	log.SetLevel(log.DebugLevel)

	cfg, err := parseConfig()
	if err != nil {
		return nil, err
	}

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("log.ParseLevel: %w", err)
	}

	log.SetLevel(level)

	return cfg, nil
}

type config struct {
	Port          int               `env:"PORT" envDefault:"4321"`
	Address       string            `env:"ADDRESS,expand" envDefault:":$PORT"`
	LogLevel      string            `env:"LOG_LEVEL" envDefault:"debug"`
	ResponsesDir  string            `env:"RESPONSES_DIR" envDefault:"./.sms_responses"`
	Ext2MIMEType  map[string]string `env:"EXTENSION_MIME_TYPE_MAP" envDefault:"txt:text/plain,json:application/json,yaml:text/yaml,xml:application/xml,html:text/html,csv:text/csv"`
	Method2Status map[string]int    `env:"METHOD_STATUS_MAP" envDefault:"DELETE:202,GET:200,PATCH:204,POST:201,PUT:204"`
}

func parseConfig() (*config, error) {
	var cfg config
	return &cfg, env.Parse(&cfg)
}
