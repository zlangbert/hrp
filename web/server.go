package web

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.nike.com/zlangb/helm-proxy/backend"
)

type context struct {
	echo.Context

	backend backend.Backend
}

func Start(backend backend.Backend) {
	e := echo.New()

	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// create custom context containing storage backend
	e.Use(appContext(backend))
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.GET("/health", health)
	e.GET("/index.yaml", index)
	e.GET("/:chart", getChart)
	e.POST("/chart", putChart)

	e.Logger.Fatal(e.Start(":1323"))
}

/*
 * custom request context to hold the backend
 */
func appContext(backend backend.Backend) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &context{
				Context: c,
				backend: backend,
			}
			return h(cc)
		}
	}
}
