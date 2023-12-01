package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/agukrapo/simpler-mock-server/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := setup()
	if err != nil {
		return err
	}

	s, err := server.New(cfg.responsesPath, cfg.contentTypePath)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	go func() {
		if err := s.Start(cfg.serverAddress); err != nil {
			log.Error("server.Start: %v", err)
		}
	}()
	defer stop(s)

	log.Infof("Server started on %s", cfg.serverAddress)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	return nil

}

func setup() (*config, error) {
	log.SetLevel(log.DebugLevel)

	cfg := parseConfig()

	level, err := log.ParseLevel(cfg.logLevel)
	if err != nil {
		return nil, fmt.Errorf("log.ParseLevel: %w", err)
	}

	log.SetLevel(level)

	return cfg, nil
}

type config struct {
	logLevel        string
	responsesPath   string
	serverAddress   string
	contentTypePath string
}

func parseConfig() *config {
	return &config{
		logLevel:        getEnv("SIMPLER-MOCK-SERVER_LOG-LEVEL", "debug"),
		responsesPath:   getEnv("SIMPLER-MOCK-SERVER_RESPONSES-PATH", "./responses"),
		serverAddress:   getEnv("SIMPLER-MOCK-SERVER_ADDRESS", ":4321"),
		contentTypePath: getEnv("SIMPLER-MOCK-SERVER_CONTENT_TYPES-PATH", "./content-type-mapping.txt"),
	}
}

func getEnv(key, fallback string) string {
	if out, ok := os.LookupEnv(key); ok {
		return out
	}

	log.Debugf("Env var %s not found, using fallback %q", key, fallback)
	return fallback
}

func stop(s *server.Server) {
	if err := s.Stop(); err != nil {
		log.Error("server.Stop: %v", err)
	}
}
