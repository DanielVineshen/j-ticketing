// FILE: cmd/core/main.go (Updated with email service)
package main

import (
	"fmt"
	authHandlers "j-ticketing/internal/core/handlers"
	coreHandlers "j-ticketing/internal/core/handlers"
	authRoutes "j-ticketing/internal/core/routes"
	coreRoutes "j-ticketing/internal/core/routes"
	service "j-ticketing/internal/core/services"
	"j-ticketing/internal/db"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/jwt"
	"j-ticketing/pkg/middleware"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
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

	// Initialize email service
	emailService := email.NewEmailService(cfg)

	// Initialize repositories
	ticketGroupRepo := repositories.NewTicketGroupRepository(database)
	bannerRepo := repositories.NewBannerRepository(database)
	adminRepo := repositories.NewAdminRepository(database)
	customerRepo := repositories.NewCustomerRepository(database)
	tokenRepo := repositories.NewTokenRepository(database)
	// auditLogRepo := repositories.NewAuditLogRepository(database)

	// Initialize services
	ticketGroupService := service.NewTicketGroupService(ticketGroupRepo, bannerRepo)
	authService := service.NewAuthService(
		jwtService,
		adminRepo,
		customerRepo,
		tokenRepo,
		emailService, // Add email service to auth service
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize handlers
	ticketGroupHandler := coreHandlers.NewTicketGroupHandler(ticketGroupService)
	authHandler := authHandlers.NewAuthHandler(authService, emailService) // Update auth handler with email service

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
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// // Apply audit logging middleware (if needed)
	// app.Use(middleware.AuditMiddleware(auditLogRepo))

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
