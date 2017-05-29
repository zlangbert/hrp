package web

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.nike.com/zlangb/hrp/config"
	"github.nike.com/zlangb/hrp/backend"
)

type context struct {
	echo.Context

	cfg *config.AppConfig
	backend backend.Backend
}

func Start(cfg *config.AppConfig, backend backend.Backend) {
	e := echo.New()

	if cfg.Debug {
		e.Debug = cfg.Debug
		e.Logger.SetLevel(log.DEBUG)
	}

	// create custom context containing config
	e.Use(appContext(cfg, backend))
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
func appContext(cfg *config.AppConfig, backend backend.Backend) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &context{
				Context: c,
				cfg: cfg,
				backend: backend,
			}
			return h(cc)
		}
	}
}
