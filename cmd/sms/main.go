package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/agukrapo/simpler-mock-server/filesystem"
	"github.com/agukrapo/simpler-mock-server/internal/mime"
	"github.com/agukrapo/simpler-mock-server/server"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if processFlags() {
		return nil
	}

	cfg, err := setup()
	if err != nil {
		return err
	}

	fs, err := filesystem.New(cfg.ResponsesDir, mime.New(cfg.Ext2MIMEType), cfg.Method2Status)
	if err != nil {
		return err
	}
	defer fs.Stop()

	s := server.New(cfg.Address, fs)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		s.Stop(ctx)
	}()

	log.Info().Msgf("Server started on %s", cfg.Address)
	defer log.Info().Msg("Server stopped")

	return s.Start(ctx)
}
