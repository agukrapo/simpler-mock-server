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
	log "github.com/sirupsen/logrus"
)

type fs interface {
	Paths() ([]*filesystem.Descriptor, error)
	Create(*http.Request) (*filesystem.Descriptor, error)
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

type dir map[string]*filesystem.Descriptor

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

func New(address string, fs fs) (*Server, error) {
	out := &Server{
		fs:     fs,
		routes: make(map[route]dir),
	}

	out.s = &http.Server{
		Addr:              address,
		Handler:           http.HandlerFunc(out.handle),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := out.refresh(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Server) Start() error {
	if err := s.s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

func (s *Server) refresh() error {
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
			log.WithFields(fieldsFromDescriptor(desc)).Warn("Route already exist")
			continue
		}

		s.routes[r][desc.Type] = desc
		log.WithFields(fieldsFromDescriptor(desc)).Debug("Route added")

		count++
	}

	if count == 0 {
		log.Warn("No routes found")
	}

	return nil
}

func (s *Server) handle(writer http.ResponseWriter, req *http.Request) {
	desc, err := s.resolveRoute(req)
	if err != nil {
		log.Errorf("Resolving route failed: %v", err)
		http.NotFound(writer, req)
		return
	}

	reader, err := desc.Reader()
	if err != nil {
		log.WithFields(fieldsFromDescriptor(desc)).Errorf("Reading route failed: %v", err)
		http.NotFound(writer, req)
		return
	}
	defer reader.Close()

	writer.Header().Set("Content-Type", desc.Type)
	writer.WriteHeader(desc.Status)

	if _, err := io.Copy(writer, reader); err != nil {
		log.WithFields(fieldsFromDescriptor(desc)).Errorf("File copy failed failed: %v", err)
		http.NotFound(writer, req)
		return
	}

	log.WithFields(log.Fields{
		"request": fmt.Sprintf("%s %s", req.Method, req.URL),
		"status":  fmt.Sprintf("%d %s", desc.Status, http.StatusText(desc.Status)),
	}).Debug("Call received")
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
		log.WithFields(fieldsFromDescriptor(desc)).Debug("Route added")
	}

	return desc, nil
}

func fieldsFromDescriptor(desc *filesystem.Descriptor) log.Fields {
	return log.Fields{
		"method": desc.Method,
		"route":  desc.Route,
		"status": desc.Status,
		"path":   desc.Path,
		"type":   desc.Type,
	}
}
