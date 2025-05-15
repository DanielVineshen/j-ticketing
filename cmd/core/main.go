// FILE: cmd/core/main.go (Updated with auth integration)
package main

import (
	"fmt"
	authHandlers "j-ticketing/internal/auth/handlers"
	"j-ticketing/internal/auth/jwt"
	authRoutes "j-ticketing/internal/auth/routes"
	"j-ticketing/internal/auth/service"
	coreHandlers "j-ticketing/internal/core/handlers"
	coreRoutes "j-ticketing/internal/core/routes"
	"j-ticketing/internal/db"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/internal/services"
	"j-ticketing/pkg/config"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Control auto-migration from environment variable
	autoMigrate := false
	if os.Getenv("AUTO_MIGRATE") == "true" {
		autoMigrate = true
		log.Println("Auto-migration is enabled")
	}

	createConstraints := false
	if os.Getenv("CREATE_CONSTRAINTS") == "true" {
		createConstraints = true
		log.Println("Create constraints is enabled")
	}

	// Initialize database connection with auto-migration control
	dbConfig := &db.DBConfig{
		AutoMigrate:       autoMigrate,
		CreateConstraints: createConstraints,
		LogLevel:          logger.Info,
	}

	database, err := db.GetDBConnection(cfg, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// If auto-migration is disabled, run explicit migrations instead
	if !autoMigrate {
		log.Println("Running explicit SQL migrations...")
		if err := db.RunMigrations(database); err != nil {
			log.Printf("Warning: Migration error: %v", err)
		}
	}

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg)

	// Initialize repositories
	ticketGroupRepo := repositories.NewTicketGroupRepository(database)
	bannerRepo := repositories.NewBannerRepository(database)
	adminRepo := repositories.NewAdminRepository(database)
	customerRepo := repositories.NewCustomerRepository(database)
	tokenRepo := repositories.NewTokenRepository(database)

	// Initialize services
	ticketGroupService := services.NewTicketGroupService(ticketGroupRepo, bannerRepo)
	authService := service.NewAuthService(
		jwtService,
		adminRepo,
		customerRepo,
		tokenRepo,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize handlers
	ticketGroupHandler := coreHandlers.NewTicketGroupHandler(ticketGroupService)
	authHandler := authHandlers.NewAuthHandler(authService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"message": "An error occurred",
				"error":   err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Setup routes
	coreRoutes.SetupRoutes(app, ticketGroupHandler, jwtService)
	authRoutes.SetupAuthRoutes(app, authHandler, jwtService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
