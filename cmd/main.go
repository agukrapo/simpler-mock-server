package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/agukrapo/simpler-mock-server/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	loggingSetup()

	rp := getEnv("SIMPLER-MOCK-SERVER_RESPONSES-PATH", "./responses")
	ctp := getEnv("SIMPLER-MOCK-SERVER_CONTENT_TYPES-PATH", "./content-type-mapping.txt")
	s, err := server.New(rp, ctp)
	if err != nil {
		log.Fatal(err)
	}

	port := getEnv("SIMPLER-MOCK-SERVER_PORT", "4321")
	addr := fmt.Sprintf(":%s", port)
	go func() {
		if err := s.Start(addr); err != nil {
			log.Fatal(err)
		}
	}()
	defer stop(s)

	log.Infof("Server started on %s", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Infof("Server shutting down due signal %q", sig)
}

func loggingSetup() {
	log.SetLevel(log.DebugLevel)

	lvl := getEnv("SIMPLER-MOCK-SERVER_LOG-LEVEL", "debug")
	level, err := log.ParseLevel(lvl)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)
}

func stop(s *server.Server) {
	if err := s.Stop(); err != nil {
		log.Error(err)
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	log.Debugf("Env var %s not found, using fallback %q", key, fallback)
	return fallback
}
