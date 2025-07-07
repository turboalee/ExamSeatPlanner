package main

import (
	"ExamSeatPlanner/internal/bootstrap"
	pkg "ExamSeatPlanner/pkg/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.uber.org/fx"
)

func main() {
	bootstrap.Loadenv()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	app := fx.New(
		pkg.EchoModules,
	)

	app.Run()
}
