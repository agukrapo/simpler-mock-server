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

	fs, err := filesystem.New(cfg.ResponsesDir, cfg.Ext2ContType, cfg.Method2Status)
	if err != nil {
		return err
	}
	defer fs.Stop()

	s := server.New(cfg.Address, fs)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := s.Start(ctx); err != nil {
			log.Error(err)
		}
		stop()
	}()

	log.Infof("Server started on %s", cfg.Address)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s.Stop(ctx)

	return nil
}
