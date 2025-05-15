package main

import (
	"fmt"
	"j-ticketing/pkg/config"
	"log"

	"j-ticketing/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	// database, err := db.GetDBConnection(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to connect to database: %v", err)
	// }

	// // Initialize repositories
	// ticketGroupRepo := repositories.NewTicketGroupRepository(database)
	// bannerRepo := repositories.NewBannerRepository(database)

	// // Initialize services
	// ticketGroupService := services.NewTicketGroupService(ticketGroupRepo, bannerRepo)

	// // Initialize handlers
	// ticketGroupHandler := handlers.NewTicketGroupHandler(ticketGroupService)

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler(logger),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())

	// Setup routes
	// routes.SetupRoutes(app, ticketGroupHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
