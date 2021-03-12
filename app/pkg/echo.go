package pkg

import (
	"context"
	"control-mitsubishi-plc-r-kube/config"
	"control-mitsubishi-plc-r-kube/kanban"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type server struct {
	Server  *echo.Echo
	Context context.Context
	Port    string
}

type Server interface {
	Start(errC chan error)
	Shutdown(ctx context.Context) error
}

func New(ctx context.Context, cfg *config.Config) Server {
	// Echo instance
	e := echo.New()
	// Routes
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// use echo default logger
	e.Use(middleware.Logger())
	e.GET("/rec/start", StartRec)
	e.GET("/rec/stop", StopRec)
	return &server{
		Server:  e,
		Context: ctx,
		Port:    fmt.Sprintf(":%v", cfg.Server.Port),
	}
}

func (s *server) Start(errC chan error) {
	err := s.Server.Start(s.Port)
	errC <- err
}

func (s *server) Shutdown(ctx context.Context) error {
	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func StartRec(c echo.Context) error {
	data := map[string]interface{}{
		"status": 0,
	}
	if err := kanban.WriteKanban(c.Request().Context(), data); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, "ok")
}

func StopRec(c echo.Context) error {
	data := map[string]interface{}{
		"status": 1,
	}
	if err := kanban.WriteKanban(c.Request().Context(), data); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, "ok")
}
