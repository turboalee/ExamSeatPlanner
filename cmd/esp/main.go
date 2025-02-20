package main

import (
	"ExamSeatPlanner/internal/bootstrap"
	pkg "ExamSeatPlanner/pkg/routes"

	"go.uber.org/fx"
)

func main() {
	bootstrap.Loadenv()
	app := fx.New(
		pkg.EchoModules,
	)

	app.Run()
}
