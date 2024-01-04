package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

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

func (c *config) extensionContentTypeMapping() map[string]string {
	out := map[string]string{
		"txt":  "text/plain",
		"json": "application/json",
		"yaml": "text/yaml",
		"xml":  "application/xml",
		"html": "text/html",
		"csv":  "text/csv",
	}

	if c.contentTypePath == "" {
		return out
	}

	f, err := os.Open(filepath.Clean(c.contentTypePath))
	if err != nil {
		log.Warnf("Unable to open content-types file: %v", err)
		return out
	}
	defer closeOrLog(f)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		chunks := strings.Split(line, "=")
		if len(chunks) != 2 {
			log.Warnf("invalid line %q", line)
			continue
		}

		out[chunks[0]] = chunks[1]
	}

	if err := scanner.Err(); err != nil {
		log.Warnf("content-type file processing failed: %v", err)
	}

	return out
}

func (c *config) methodStatusMapping() map[string]int {
	out := map[string]int{
		http.MethodDelete: http.StatusAccepted,
		http.MethodGet:    http.StatusOK,
		http.MethodPatch:  http.StatusNoContent,
		http.MethodPost:   http.StatusCreated,
		http.MethodPut:    http.StatusNoContent,
	}

	return out
}

func closeOrLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Error(err)
	}
}
