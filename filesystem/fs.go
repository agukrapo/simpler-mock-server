package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	root         string
	contentTypes map[string]string
	methods      map[string]int
}

func New(root string, contentTypes map[string]string, methods map[string]int) (*FS, error) {
	if err := validate(root); err != nil {
		return nil, err
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &FS{
		root:         root,
		contentTypes: contentTypes,
		methods:      methods,
	}, nil
}

func (fs *FS) Paths() ([]*Descriptor, error) {
	var out []*Descriptor

	for method, status := range fs.methods {
		sp, err := fs.subPaths(method, status)
		if err != nil {
			log.WithFields(log.Fields{
				"method": method,
				"status": status,
				"error":  err,
			}).Warnf("Unable to process paths")

			continue
		}

		out = append(out, sp...)
	}

	return out, nil
}

func (fs *FS) subPaths(method string, status int) ([]*Descriptor, error) {
	root := filepath.Join(fs.root, method)
	if err := validate(root); err != nil {
		return nil, err
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
			return err
		}

		ct, ok := fs.contentTypes[ext]
		if !ok {
			return fmt.Errorf("content-type not found for extension %q", ext)
		}

		name, status, err := parseStatus(name, status)
		if err != nil {
			return fmt.Errorf("parseStatus: %w", err)
		}

		out = append(out, &Descriptor{
			Method: method,
			Path:   path,
			Route:  dir + name,
			Status: status,
			Type:   ct,
			Reader: func() (io.ReadCloser, error) {
				return os.Open(path)
			},
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.Walk: %w", err)
	}

	return out, nil
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

func splitBase(base string) (string, string, error) {
	chunks := strings.Split(base, ".")
	if len(chunks) != 2 {
		return "", "", fmt.Errorf("invalid file %q", base)
	}

	return chunks[0], chunks[1], nil
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
