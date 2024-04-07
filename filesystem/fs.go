package filesystem

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agukrapo/simpler-mock-server/internal/headers"
	"github.com/agukrapo/simpler-mock-server/internal/mime"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type Descriptor struct {
	Method string
	Path   string
	Route  string
	Status int
	Type   mime.Type
	Reader func() (io.ReadCloser, error)
}

type FS struct {
	root string

	watcher *fsnotify.Watcher
	paths   map[string]chan fsnotify.Event
	mu      sync.Mutex
	events  chan struct{}

	types *mime.Types

	method2Status map[string]int
}

func New(root string, types *mime.Types, method2Status map[string]int) (*FS, error) {
	if err := validate(root); err != nil {
		return nil, err
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FS{
		root:          root,
		watcher:       watcher,
		events:        make(chan struct{}),
		types:         types,
		method2Status: method2Status,
	}, nil
}

func (fs *FS) Stop() {
	if err := fs.watcher.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close filesystem watcher")
	}
	close(fs.events)
}

func (fs *FS) Paths() ([]*Descriptor, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := validate(fs.root); err != nil {
		return nil, err
	}

	fs.resetWatcher()

	if err := fs.watcher.Add(fs.root); err != nil {
		log.Error().Err(err).Msgf("Failed to watch %s dir", fs.root)
	}

	var out []*Descriptor

	for method, status := range fs.method2Status {
		sp, err := fs.subPaths(method, status)
		if err != nil {
			log.Error().Str("method", method).Int("status", status).Err(err).Msg("Failed to process paths")

			continue
		}

		out = append(out, sp...)
	}

	go fs.eventLoop()

	return out, nil
}

func (fs *FS) subPaths(method string, status int) ([]*Descriptor, error) {
	root := filepath.Clean(filepath.Join(fs.root, method))

	if _, err := os.Open(root); os.IsNotExist(err) {
		return nil, nil
	}

	var out []*Descriptor

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsDir() {
			if err := fs.watcher.Add(path); err != nil {
				log.Error().Err(err).Msgf("Failed to watch %s dir", path)
			}
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		dir, base := filepath.Split(path)
		dir = strings.TrimPrefix(dir, filepath.Join(fs.root, method))

		name, ext, err := splitBase(base)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to split path %s", base)
			return nil
		}

		name, status, err := parseStatus(name, status)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse status from name %s", name)
			return nil
		}

		out = append(out, &Descriptor{
			Method: method,
			Path:   strings.TrimPrefix(path, filepath.Dir(fs.root)+"/"),
			Route:  dir + name,
			Status: status,
			Type:   fs.types.Type(mime.Extension(ext)),
			Reader: func() (io.ReadCloser, error) {
				return os.Open(filepath.Clean(path))
			},
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.Walk: %w", err)
	}

	return out, nil
}

func (fs *FS) eventLoop() {
	t := time.AfterFunc(math.MaxInt64, func() {
		select {
		case fs.events <- struct{}{}:
		default:
		}
	})
	t.Stop()

	for {
		select {
		case event, ok := <-fs.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				t.Reset(200 * time.Millisecond)
			}
		case err, ok := <-fs.watcher.Errors:
			if !ok {
				return
			}

			log.Error().Err(err).Msg("Filesystem watcher failed")
		}
	}
}

func (fs *FS) Create(req *http.Request) (*Descriptor, error) {
	ext := fs.types.Extension(headers.Accept(req))
	path := strings.TrimSuffix(req.URL.Path, "/")
	file := filepath.Clean(filepath.Join(fs.root, req.Method, fmt.Sprintf("%s.%s", path, ext)))

	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return nil, fmt.Errorf("os.MkdirAll: %w", err)
	}

	if _, err := os.Create(file); err != nil {
		return nil, fmt.Errorf("os.Create: %w", err)
	}

	return &Descriptor{
		Method: req.Method,
		Path:   file,
		Route:  strings.TrimPrefix(path, filepath.Dir(fs.root)+"/"),
		Status: fs.method2Status[req.Method],
		Type:   fs.types.Type(ext),
		Reader: func() (io.ReadCloser, error) {
			return os.Open(file)
		},
	}, nil
}

func validate(dir string) error {
	f, err := os.Open(filepath.Clean(dir))
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}

func splitBase(path string) (string, string, error) {
	if path == "" {
		return "", "", errors.New("input is empty")
	}

	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext), strings.TrimPrefix(ext, "."), nil
}

func parseStatus(name string, status int) (string, int, error) {
	chunks := strings.Split(name, "___")
	if len(chunks) != 2 {
		return name, status, nil
	}

	ns, err := strconv.Atoi(chunks[0])
	if err != nil {
		return "", 0, fmt.Errorf("invalid status %q: %w", chunks[0], err)
	}

	return chunks[1], ns, nil
}

func (fs *FS) Notify() <-chan struct{} {
	return fs.events
}

func (fs *FS) resetWatcher() {
	for p := range fs.paths {
		if err := fs.watcher.Remove(p); err != nil {
			log.Error().Err(err).Msgf("Filed to unwatch path %v", p)
		}
	}
	clear(fs.paths)
}
