package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/agukrapo/simpler-mock-server/filesystem"
	"github.com/labstack/echo/v4"
	glog "github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
)

type fs interface {
	Paths() ([]*filesystem.Descriptor, error)
}

type Server struct {
	echo *echo.Echo
}

func New(fs fs) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(glog.OFF)
	e.Use(logCalls)

	out := &Server{
		echo: e,
	}

	paths, err := fs.Paths()
	if err != nil {
		return nil, err
	}

	var count uint8
	for _, desc := range paths {
		if err := out.addRoute(desc); err != nil {
			return nil, err
		}
		count++
	}

	if count == 0 {
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

func (s *Server) Stop(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func (s *Server) addRoute(desc *filesystem.Descriptor) error {
	handler := func(c echo.Context) error {
		f, err := desc.Reader()
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		return c.Blob(desc.Status, desc.Type, b)
	}

	s.echo.Add(desc.Method, desc.Route, handler)
	s.echo.Add(desc.Method, desc.Route+"/", handler)

	log.WithFields(log.Fields{
		"method": desc.Method,
		"route":  desc.Route,
		"status": desc.Status,
		"path":   desc.Path,
	}).Debug("Route added")

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
