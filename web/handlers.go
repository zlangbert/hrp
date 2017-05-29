package web

import (
	"github.com/labstack/echo"
	"net/http"
)

func health(c echo.Context) error {
	return c.NoContent(200)
}

func index(ec echo.Context) error {
	c := ec.(*context)
	index, err := c.backend.GetIndex()
	if err != nil {
		return err
	}

	return c.Blob(200, "text/yaml", index)
}

func getChart(ec echo.Context) error {
	c := ec.(*context)

	name := c.Param("chart")
	chart, err := c.backend.GetChart(name)
	if err != nil {
		return err
	}

	return c.Blob(200, "application/gzip", chart)
}

func putChart(ec echo.Context) error {
	c := ec.(*context)

	file, err := c.FormFile("chart")
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "missing 'chart' param")
	}

	c.Logger().Infof("putting charts %s", file.Filename)

	err = c.backend.PutChart(file)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"backend failed put chart")
	}

	return c.NoContent(200)
}

func reindex(ec echo.Context) error {
	c := ec.(*context)

	err := c.backend.Reindex()
	if err != nil {
		return echo.NewHTTPError(500, "failed to reindex")
	}

	return c.NoContent(200)
}
