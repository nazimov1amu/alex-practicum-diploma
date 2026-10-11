package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Handler interface {
	Register(g *echo.Group)
}

func Router(handlers []Handler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())

	api := e.Group("/api")
	api.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	for _, h := range handlers {
		h.Register(api)
	}

	return e
}
