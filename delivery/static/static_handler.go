package static

import (
	"github.com/labstack/echo/v4"
)

func RegisterStaticRoutes(e *echo.Echo) {
	// Serve static files
	e.Static("/", "static")
	e.GET("/", func(c echo.Context) error {
		return c.File("views/index.html")
	})
}
