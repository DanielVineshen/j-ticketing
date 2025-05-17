// File: j-ticketing/cmd/core/main.go
package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"j-ticketing/internal/core/dto/payment"
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/routes"
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
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using default or environment values")
	}

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

	// Initialize payment config
	paymentConfig := payment.PaymentConfig{
		GatewayURL:      getRequiredEnv("PAYMENT_GATEWAY_URL"),
		APIKey:          getRequiredEnv("JP_API_KEY"),
		BaseURL:         getRequiredEnv("BASE_URL"),
		AGToken:         getRequiredEnv("AG_TOKEN"),
		FrontendBaseURL: getRequiredEnv("FRONTEND_BASE_URL"),
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
	adminRepo := repositories.NewAdminRepository(database)
	customerRepo := repositories.NewCustomerRepository(database)
	tokenRepo := repositories.NewTokenRepository(database)
	tagRepo := repositories.NewTagRepository(database)
	groupGalleryRepo := repositories.NewGroupGalleryRepository(database)
	ticketDetailRepo := repositories.NewTicketDetailRepository(database)
	orderTicketGroupRepo := repositories.NewOrderTicketGroupRepository(database)
	orderTicketInfoRepo := repositories.NewOrderTicketInfoRepository(database)

	// Initialize services
	ticketGroupService := service.NewTicketGroupService(
		ticketGroupRepo,
		tagRepo,
		groupGalleryRepo,
		ticketDetailRepo,
		cfg,
	)
	authService := service.NewAuthService(
		jwtService,
		adminRepo,
		customerRepo,
		tokenRepo,
		emailService, // Add email service to auth service
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	orderService := service.NewOrderService(
		orderTicketGroupRepo,
		orderTicketInfoRepo,
		ticketGroupRepo,
		tagRepo,
		groupGalleryRepo,
		ticketDetailRepo,
		&paymentConfig,
		ticketGroupService,
	)
	customerService := service.NewCustomerService(customerRepo)

	// Initialize handlers
	ticketGroupHandler := handlers.NewTicketGroupHandler(ticketGroupService)
	authHandler := handlers.NewAuthHandler(authService, emailService) // Update auth handler with email service
	orderHandler := handlers.NewOrderHandler(orderService, customerService)

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
	app.Static("/public", "./pkg/public")

	// // Apply audit logging middleware (if needed)
	// app.Use(middleware.AuditMiddleware(auditLogRepo))

	// Setup routes
	routes.SetupTicketGroupRoutes(app, ticketGroupHandler, jwtService)
	routes.SetupAuthRoutes(app, authHandler, jwtService)
	routes.SetupOrderRoutes(app, orderHandler, jwtService)
	routes.SetupPaymentRoutes(app, paymentConfig, orderTicketGroupRepo, orderTicketInfoRepo)
	routes.SetupViewRoutes(app)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.CorePort)
	log.Printf("Server starting on %s with version 0.1", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// Helper function to get environment variables with required check
func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Environment variable %s is required but not set", key)
	}
	return value
}
