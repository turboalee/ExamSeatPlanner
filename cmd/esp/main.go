package main

import (
	"ExamSeatPlanner/internal/bootstrap"
	"ExamSeatPlanner/internal/config"
	"log"

	"go.uber.org/fx"
)

func main() {
	bootstrap.Loadenv()
	app := fx.New(
		config.MongoModule,
		fx.Invoke(func(client *config.MongoDBClient) {
			userCollection := client.GetCollection("users")
			log.Println("User collection ready:", userCollection.Name())
		}),
	)

	app.Run()
}
