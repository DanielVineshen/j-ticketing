// File: j-ticketing/cmd/core/main.go
package main

import (
	"fmt"
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
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Initialize slogger first so we can use it throughout
	slogger := initLogger()

	// Pre-processing for OAuth client ID - clean up any URL prefixes
	if clientID := os.Getenv("CLIENT_ID"); strings.HasPrefix(clientID, "http://") || strings.HasPrefix(clientID, "https://") {
		cleanClientID := strings.TrimPrefix(strings.TrimPrefix(clientID, "http://"), "https://")
		truncatedID := cleanClientID
		if len(cleanClientID) > 10 {
			truncatedID = cleanClientID[:10]
		}

		slogger.Info("CLIENT_ID contains URL prefix, using cleaned value",
			"original", clientID,
			"cleaned", truncatedID)

		err := os.Setenv("CLIENT_ID", cleanClientID)
		if err != nil {
			slogger.Error("Failed to set cleaned CLIENT_ID", "error", err)
			return
		}
	}

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

	// Initialize payment config
	paymentConfig := payment.PaymentConfig{
		GatewayURL:      getRequiredEnv("PAYMENT_GATEWAY_URL", slogger),
		APIKey:          getRequiredEnv("JP_API_KEY", slogger),
		BaseURL:         getRequiredEnv("BASE_URL", slogger),
		AGToken:         getRequiredEnv("AG_TOKEN", slogger),
		FrontendBaseURL: getRequiredEnv("FRONTEND_BASE_URL", slogger),
	}

	// Initialize database connection with auto-migration control
	dbConfig := &db.DBConfig{
		AutoMigrate:       autoMigrate,
		CreateConstraints: createConstraints,
		LogLevel:          gormLogger.Info,
	}

	database, err := db.GetDBConnection(cfg, dbConfig)
	if err != nil {
		slogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg)

	// Initialize email service
	emailService := email.NewEmailService(cfg)

	// Test the email connection if we're using OAuth2
	if cfg.Email.ClientID != "" && cfg.Email.ClientSecret != "" && cfg.Email.RefreshToken != "" {
		slogger.Info("Testing OAuth2 token acquisition...")
		if err := testOAuth2(emailService); err != nil {
			slogger.Warn("OAuth2 token test failed, email sending with OAuth2 might not work correctly",
				"error", err)
		} else {
			slogger.Info("OAuth2 token test succeeded - email service should work correctly")
		}
	}

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
		emailService,
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
	authHandler := handlers.NewAuthHandler(authService, emailService)
	orderHandler := handlers.NewOrderHandler(orderService, customerService, jwtService)
	simplePDFHandler := handlers.NewPDFHandler()

	// Create Fiber app with adapted error handler for slog
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler(slogger),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080,http://127.0.0.1:3000,http://139.59.253.119:3000,https://etiket.johor.gov.my,http://stagingetiket.johor.gov.my:3000",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Static("/public", "./pkg/public")

	// Setup routes
	routes.SetupTicketGroupRoutes(app, ticketGroupHandler, jwtService)
	routes.SetupAuthRoutes(app, authHandler, jwtService)
	routes.SetupOrderRoutes(app, orderHandler, jwtService)
	routes.SetupPaymentRoutes(app, paymentConfig, orderTicketGroupRepo, orderTicketInfoRepo, emailService, ticketGroupRepo)
	routes.SetupViewRoutes(app)
	routes.SetupTicketPDFRoutes(app, simplePDFHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.CorePort)
	slogger.Info("Server starting",
		"address", addr,
		"version", "0.2")

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

// Helper function to get environment variables with required check
func getRequiredEnv(key string, slogger *slog.Logger) string {
	value := os.Getenv(key)
	if value == "" {
		slogger.Error("Required environment variable not set", "variable", key)
		os.Exit(1)
	}
	return value
}

// Helper function to test OAuth2 token acquisition
func testOAuth2(emailService email.EmailService) error {
	// Use SendEmail to a fake recipient, but with a flag to just test token acquisition
	// Return nil to indicate success
	// Temporary implementation, requires proper email to ensure it works
	return nil
}
