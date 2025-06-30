package pkg

import (
	"ExamSeatPlanner/internal/auth"
	"ExamSeatPlanner/internal/config"
	"ExamSeatPlanner/internal/notification"
	"ExamSeatPlanner/internal/seating"
	"ExamSeatPlanner/pkg/middleware"
	"context"
	"log"
	"os"

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
	fx.Provide(notification.NewNotificationRepository),
	fx.Provide(notification.NewNotificationService),
	fx.Provide(notification.NewNotificationHandler),
	fx.Provide(notification.NewNotificationScheduler),
	fx.Provide(seating.NewSeatingRepository),
	fx.Provide(seating.NewSeatingService),
	fx.Provide(seating.NewSeatingHandler),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(StartNotificationScheduler))

func NewEchoServer(lc fx.Lifecycle) *echo.Echo {
	e := echo.New()
	middleware.SetupMiddleware(e)
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // Default port if not specified in environment
	}
	if port[0] != ':' {
		port = ":" + port
	}
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

// StartNotificationScheduler starts the notification scheduler using dependency injection.
func StartNotificationScheduler(scheduler *notification.NotificationScheduler, lc fx.Lifecycle) {
	scheduler.StartScheduler(lc)
}

func RegisterRoutes(e *echo.Echo, authHandler *auth.AuthHandler, notificationHandler *notification.NotificationHandler, seatingHandler *seating.SeatingHandler) {
	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)
	e.POST("/forgot-Password", authHandler.ForgotPassword)
	e.POST("/verify-email", authHandler.VerifyEmail)
	e.POST("/reset-password", authHandler.ResetPassword)

	protected := e.Group("/api")
	protected.Use(middleware.JWTMiddleware)
	protected.Use(middleware.CasbinMiddleware)
	protected.GET("/profile", authHandler.Profile)

	// Notification routes (admin only)
	protected.POST("/notifications/schedule", notificationHandler.ScheduleNotification)

	// Seating routes
	seating := protected.Group("/seating")
	seating.POST("/generate", seatingHandler.GenerateSeatingPlan)   // Admin only
	seating.GET("/plans/:id", seatingHandler.GetSeatingPlan)        // All authenticated users
	seating.POST("/exams", seatingHandler.CreateExam)               // Admin only
	seating.POST("/rooms", seatingHandler.CreateRoom)               // Admin only
	seating.POST("/students", seatingHandler.CreateStudent)         // Staff only
	seating.POST("/invigilators", seatingHandler.CreateInvigilator) // Admin only
}
