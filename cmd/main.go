package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/agukrapo/simpler-mock-server/filesystem"
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

	fs, err := filesystem.New(cfg.ResponsesPath, cfg.ContentTypeMap, cfg.MethodStatusMap)
	if err != nil {
		return fmt.Errorf("filesystem.New: %w", err)
	}

	s, err := server.New(fs)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	go func() {
		if err := s.Start(cfg.ServerAddress); err != nil {
			log.Errorf("server.Start: %v", err)
		}
	}()

	log.Infof("Server started on %s", cfg.ServerAddress)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		log.Errorf("server.Stop: %v", err)
	}

	return nil
}
