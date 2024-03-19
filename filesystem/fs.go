package filesystem

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/agukrapo/simpler-mock-server/internal/bimap"
	"github.com/agukrapo/simpler-mock-server/internal/headers"
	log "github.com/sirupsen/logrus"
)

type Descriptor struct {
	Method string
	Path   string
	Route  string
	Status int
	Type   string
	Reader func() (io.ReadCloser, error)
}

type FS struct {
	root string

	ext2ContType  *bimap.Bimap[string, string]
	method2Status map[string]int
}

func New(root string, ext2ContType map[string]string, method2Status map[string]int) (*FS, error) {
	if err := validate(root); err != nil {
		return nil, err
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs: %w", err)
	}

	return &FS{
		root:          root,
		ext2ContType:  bimap.New[string, string](ext2ContType),
		method2Status: method2Status,
	}, nil
}

func (fs *FS) Paths() ([]*Descriptor, error) {
	var out []*Descriptor

	for method, status := range fs.method2Status {
		sp, err := fs.subPaths(method, status)
		if err != nil {
			log.WithFields(log.Fields{
				"method": method,
				"status": status,
				"error":  err,
			}).Errorf("Unable to process paths")

			continue
		}

		out = append(out, sp...)
	}

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

		if !info.Mode().IsRegular() {
			return nil
		}

		dir, base := filepath.Split(path)
		dir = strings.TrimPrefix(dir, filepath.Join(fs.root, method))

		name, ext, err := splitBase(base)
		if err != nil {
			log.Error(err)
			return nil
		}

		ct, ok := fs.ext2ContType.GetByKey(ext)
		if !ok {
			log.Errorf("content-type not found for extension %q", ext)
			return nil
		}

		name, status, err := parseStatus(name, status)
		if err != nil {
			log.Error(err)
			return nil
		}

		out = append(out, &Descriptor{
			Method: method,
			Path:   path,
			Route:  dir + name,
			Status: status,
			Type:   ct,
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

func (fs *FS) Create(req *http.Request) (*Descriptor, error) {
	ext := "json"
	if ct := headers.Accept(req); ct != "" {
		if e, ok := fs.ext2ContType.GetByValue(ct); ok {
			ext = e
		}
	}

	path := strings.TrimSuffix(req.URL.Path, "/")
	file := filepath.Clean(filepath.Join(fs.root, req.Method, path+"."+ext))

	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return nil, fmt.Errorf("os.MkdirAll: %w", err)
	}

	if _, err := os.Create(file); err != nil {
		return nil, fmt.Errorf("os.Create: %w", err)
	}

	ct, _ := fs.ext2ContType.GetByKey(ext)

	return &Descriptor{
		Method: req.Method,
		Path:   file,
		Route:  path,
		Status: fs.method2Status[req.Method],
		Type:   ct,
		Reader: func() (io.ReadCloser, error) {
			return os.Open(file)
		},
	}, nil
}

func validate(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
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
