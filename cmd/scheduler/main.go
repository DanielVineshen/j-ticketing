// File: j-ticketing/cmd/scheduler/main.go
package main

import (
	"fmt"
	"j-ticketing/internal/db"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/middleware"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Initialize slogger first so we can use it throughout
	slogger := initLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slogger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Control auto-migration from environment variable
	autoMigrate := false
	if os.Getenv("AUTO_MIGRATE") == "true" {
		autoMigrate = true
		slogger.Info("Auto-migration is enabled")
	}

	createConstraints := false
	if os.Getenv("CREATE_CONSTRAINTS") == "true" {
		createConstraints = true
		slogger.Info("Create constraints is enabled")
	}

	// Initialize database connection with auto-migration control
	dbConfig := &db.DBConfig{
		AutoMigrate:       autoMigrate,
		CreateConstraints: createConstraints,
		LogLevel:          gormLogger.Info,
	}

	_, err = db.GetDBConnection(cfg, dbConfig)
	if err != nil {
		slogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler(slogger),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.SchedulerPort)
	slogger.Info("Server starting",
		"address", addr,
		"version", "0.1")

	if err := app.Listen(addr); err != nil {
		slogger.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

// initLogger sets up an slog slogger with appropriate configuration
func initLogger() *slog.Logger {
	// Determine if we're in development mode
	isDev := os.Getenv("APP_ENV") != "production"

	var handler slog.Handler
	if isDev {
		// Development slogger: Text format with source location
		opts := &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Production slogger: JSON format
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
			// Custom time format similar to ISO8601
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format(time.RFC3339))
					}
				}
				return a
			},
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slogger := slog.New(handler)

	// Set as default slogger too (optional)
	slog.SetDefault(slogger)

	return slogger
}
