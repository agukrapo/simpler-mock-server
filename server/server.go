package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/agukrapo/simpler-mock-server/filesystem"
	"github.com/agukrapo/simpler-mock-server/internal/headers"
	"github.com/agukrapo/simpler-mock-server/internal/mime"
	"github.com/rs/zerolog/log"
)

type fs interface {
	Paths() ([]*filesystem.Descriptor, error)
	Create(*http.Request) (*filesystem.Descriptor, error)
	Notify() <-chan struct{}
}

type route struct {
	method, path string
}

func descriptorToRoute(desc *filesystem.Descriptor) route {
	return route{
		method: desc.Method,
		path:   desc.Route,
	}
}

func requestToRoute(req *http.Request) route {
	return route{
		method: req.Method,
		path:   req.URL.Path,
	}
}

type dir map[mime.Type]*filesystem.Descriptor

func (d dir) resolveDescriptor(req *http.Request) (*filesystem.Descriptor, bool) {
	if len(d) == 0 {
		return nil, false
	}

	ct := headers.Accept(req)
	if ct == "" {
		for _, v := range d {
			return v, true
		}
	}

	out, ok := d[ct]
	return out, ok
}

type Server struct {
	s  *http.Server
	fs fs

	routes map[route]dir
	mu     sync.RWMutex
}

func New(address string, fs fs) *Server {
	out := &Server{
		fs:     fs,
		routes: make(map[route]dir),
	}

	out.s = &http.Server{
		Addr:              address,
		Handler:           http.HandlerFunc(out.handle),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return out
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.refresh(); err != nil {
		return err
	}

	go s.watch(ctx)

	if err := s.s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) {
	if err := s.s.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}
}

func (s *Server) watch(ctx context.Context) {
	c := s.fs.Notify()
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-c:
			if !ok {
				return
			}

			if err := s.refresh(); err != nil {
				log.Error().Err(err).Msg("Routes refresh failed")
				s.Stop(ctx)
			}
		}
	}
}

func (s *Server) refresh() error {
	log.Info().Msg("Refreshing routes...")
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.routes)

	paths, err := s.fs.Paths()
	if err != nil {
		return err
	}

	var count uint8
	for _, desc := range paths {
		r := descriptorToRoute(desc)
		if _, ok := s.routes[r]; !ok {
			s.routes[r] = make(dir)
		}

		if _, ok := s.routes[r][desc.Type]; ok {
			log.Warn().Fields(fieldsFromDescriptor(desc)).Msg("Route already exist")
			continue
		}

		s.routes[r][desc.Type] = desc
		log.Debug().Fields(fieldsFromDescriptor(desc)).Msg("Route added")

		count++
	}

	if count == 0 {
		log.Warn().Msg("No routes found")
	}

	return nil
}

func (s *Server) handle(writer http.ResponseWriter, req *http.Request) {
	desc, err := s.resolveRoute(req)
	if err != nil {
		log.Error().Err(err).Msg("Resolving route failed")
		http.NotFound(writer, req)
		return
	}

	reader, err := desc.Reader()
	if err != nil {
		log.Error().Err(err).Fields(fieldsFromDescriptor(desc)).Msg("Reading route failed")
		http.NotFound(writer, req)
		return
	}
	defer reader.Close()

	writer.Header().Set("Content-Type", string(desc.Type))
	writer.WriteHeader(desc.Status)

	if _, err := io.Copy(writer, reader); err != nil {
		log.Error().Fields(fieldsFromDescriptor(desc)).Msg("File copy failed")
		http.NotFound(writer, req)
		return
	}

	log.Debug().Str("request", fmt.Sprintf("%s %s", req.Method, req.URL)).Str("status", fmt.Sprintf("%d %s", desc.Status, http.StatusText(desc.Status))).Msg("Call received")
}

func (s *Server) resolveRoute(req *http.Request) (*filesystem.Descriptor, error) {
	r := requestToRoute(req)
	desc, ok := s.routes[r].resolveDescriptor(req)
	if !ok {
		s.mu.Lock()
		defer s.mu.Unlock()

		var err error
		desc, err = s.fs.Create(req.Clone(context.Background()))
		if err != nil {
			return nil, err
		}

		if s.routes[r] == nil {
			s.routes[r] = make(dir)
		}

		s.routes[r][desc.Type] = desc
		log.Debug().Fields(fieldsFromDescriptor(desc)).Msg("Route added")
	}

	return desc, nil
}

func fieldsFromDescriptor(desc *filesystem.Descriptor) map[string]interface{} {
	return map[string]interface{}{
		"method": desc.Method,
		"route":  desc.Route,
		"status": desc.Status,
		"path":   desc.Path,
		"type":   desc.Type,
	}
}
