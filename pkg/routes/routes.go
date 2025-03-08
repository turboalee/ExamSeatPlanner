package pkg

import (
	"ExamSeatPlanner/internal/auth"
	"ExamSeatPlanner/internal/config"
	"ExamSeatPlanner/pkg/middleware"
	"context"
	"log"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

var EchoModules = fx.Module("echo",
	fx.Provide(NewEchoServer),
	fx.Provide(config.NewMongoDBConfig),
	fx.Provide(config.NewMongoDBClient),
	fx.Provide(config.NewResendConfig),
	fx.Provide(config.NewEmailService),
	fx.Provide(auth.NewUserRepository),
	fx.Provide(auth.NewAuthService),
	fx.Provide(auth.NewUserService),
	fx.Provide(auth.NewAuthHandler),
	fx.Invoke(RegisterRoutes))

func NewEchoServer(lc fx.Lifecycle) *echo.Echo {
	e := echo.New()
	middleware.SetupMiddleware(e)
	port := ":8080"
	log.Println("Server running on http://localhost" + port[1:])
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(port); err != nil {
					log.Fatal("Failed to start the server:", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("shutting down the server ...")
			return e.Shutdown(ctx)
		},
	})
	return e
}

func RegisterRoutes(e *echo.Echo, authHandler *auth.AuthHandler) {
	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)
	e.POST("/forgot-Password", authHandler.ForgotPassword)
	e.POST("/verify-email", authHandler.VerifyEmail)
	e.POST("/reset-password", authHandler.ResetPassword)

	protected := e.Group("/api")
	protected.Use(middleware.JWTMiddleware)
	protected.GET("/profile", authHandler.Profile)
}
