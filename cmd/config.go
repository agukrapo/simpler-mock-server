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
	LogLevel        string            `env:"LOG_LEVEL" envDefault:"debug"`
	ResponsesPath   string            `env:"RESPONSES_PATH" envDefault:"./responses"`
	ServerAddress   string            `env:"SERVER_ADDRESS" envDefault:":4321"`
	ContentTypeMap  map[string]string `env:"EXTENSION_CONTENT_TYPE_MAP" envDefault:"txt:text/plain,json:application/json,yaml:text/yaml,xml:application/xml,html:text/html,csv:text/csv"`
	MethodStatusMap map[string]int    `env:"METHOD_STATUS_MAP" envDefault:"DELETE:202,GET:200,PATCH:204,POST:201,PUT:204"`
}

func parseConfig() (*config, error) {
	var cfg config
	return &cfg, env.Parse(&cfg)
}
