package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	glog "github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	echo         *echo.Echo
	contentTypes map[string]string
	resPath      string
	routes       map[string]struct{}
}

func New(dir, mappingsPath string) (*Server, error) {
	if err := validate(dir); err != nil {
		return nil, err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(glog.OFF)
	e.Use(logCalls)

	m, err := contentTypeMapping(mappingsPath)
	if err != nil {
		return nil, err
	}

	out := &Server{
		echo:         e,
		contentTypes: m,
		resPath:      dir,
		routes:       make(map[string]struct{}),
	}

	if err := out.addRoutes(http.MethodDelete, http.StatusAccepted); err != nil {
		return nil, err
	}
	if err := out.addRoutes(http.MethodGet, http.StatusOK); err != nil {
		return nil, err
	}
	if err := out.addRoutes(http.MethodPatch, http.StatusNoContent); err != nil {
		return nil, err
	}
	if err := out.addRoutes(http.MethodPost, http.StatusCreated); err != nil {
		return nil, err
	}
	if err := out.addRoutes(http.MethodPut, http.StatusNoContent); err != nil {
		return nil, err
	}

	if len(out.routes) == 0 {
		return nil, errors.New("no routes found")
	}

	return out, nil
}

func (s *Server) Start(addr string) error {
	if err := s.echo.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.echo.Shutdown(ctx)
}

func (s *Server) addRoutes(method string, status int) error {
	gf, err := s.paths(method)
	if err != nil {
		return err
	}

	for _, f := range gf {
		if err := s.add(method, status, f); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) add(method string, status int, file string) error {
	dir, base := filepath.Split(file)
	dir = strings.TrimPrefix(dir, "responses/"+method)

	name, ext, err := splitBase(base)
	if err != nil {
		return fmt.Errorf("splitBase: %w", err)
	}

	ct, ok := s.contentTypes[ext]
	if !ok {
		return fmt.Errorf("content-type not found for extension %q", ext)
	}

	name, status, err = parseStatus(name, status)
	if err != nil {
		return fmt.Errorf("parseStatus: %w", err)
	}

	route := dir + name

	fn := func(c echo.Context) error {
		f, err := os.Open(filepath.Clean(file))
		if err != nil {
			return fmt.Errorf("os.Open: %w", err)
		}
		defer CloseAndLog(f)

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		return c.Blob(status, ct, b)
	}

	s.echo.Add(method, route, fn)
	s.echo.Add(method, route+"/", fn)

	s.routes[route] = struct{}{}

	log.WithFields(log.Fields{
		"method": method,
		"route":  route,
		"file":   file,
		"status": status,
	}).Debug("Route added")

	return nil
}

func (s *Server) paths(dir string) ([]string, error) {
	var out []string

	root := filepath.Join(s.resPath, dir)
	if err := validate(root); err != nil {
		log.Warn(err)
		return out, nil
	}

	err := filepath.Walk(root, func(p string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			out = append(out, p)
		}

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

func logCalls(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if err = next(c); err != nil {
			c.Error(err)
		}

		req := c.Request()
		res := c.Response()
		fields := log.Fields{
			"request": fmt.Sprintf("%s %s", req.Method, req.URL),
			"status":  fmt.Sprintf("%d %s", res.Status, http.StatusText(res.Status)),
		}
		if err != nil {
			fields["error"] = err.Error()
		}
		log.WithFields(fields).Debug("Call received")

		return err
	}
}

func contentTypeMapping(path string) (map[string]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(f)

	scanner := bufio.NewScanner(f)

	out := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()

		chunks := strings.Split(line, "=")
		if len(chunks) != 2 {
			return nil, fmt.Errorf("invalid line %q", line)
		}

		out[chunks[0]] = chunks[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func CloseAndLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Error(err)
	}
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
