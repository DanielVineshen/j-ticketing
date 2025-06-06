// File: j-ticketing/cmd/core/main.go
package main

import (
	"fmt"
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

	// Initialize payment config
	//paymentConfig := payment.PaymentConfig{
	//	GatewayURL:      getRequiredEnv("PAYMENT_GATEWAY_URL", slogger),
	//	APIKey:          getRequiredEnv("JP_API_KEY", slogger),
	//	BaseURL:         getRequiredEnv("BASE_URL", slogger),
	//	AGToken:         getRequiredEnv("AG_TOKEN", slogger),
	//	FrontendBaseURL: getRequiredEnv("FRONTEND_BASE_URL", slogger),
	//}

	// Initialize database connection with auto-migration control
	dbConfig := &db.DBConfig{
		AutoMigrate:       autoMigrate,
		CreateConstraints: createConstraints,
		LogLevel:          gormLogger.Error,
	}

	database, err := db.GetDBConnection(cfg, dbConfig)
	if err != nil {
		slogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg)

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
	bannerRepo := repositories.NewBannerRepository(database)
	orderTicketLogRepo := repositories.NewOrderTicketLogRepository(database)
	customerLogRepo := repositories.NewCustomerLogRepository(database)
	notificationRepo := repositories.NewNotificationRepository(database)
	ticketVariantRepo := repositories.NewTicketVariantRepository(database)
	generalRepo := repositories.NewGeneralRepository(database)

	// Initialize email service
	emailService := email.NewEmailService(generalRepo)

	// Initialize services
	paymentService := service.NewPaymentService(
		orderTicketGroupRepo,
		orderTicketInfoRepo,
		ticketGroupRepo,
		tagRepo,
		groupGalleryRepo,
		ticketDetailRepo,
		generalRepo,
	)
	ticketGroupService := service.NewTicketGroupService(
		ticketGroupRepo,
		tagRepo,
		groupGalleryRepo,
		ticketDetailRepo,
		ticketVariantRepo,
		generalRepo,
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
	customerService := service.NewCustomerService(customerRepo, customerLogRepo)
	orderService := service.NewOrderService(
		orderTicketGroupRepo,
		orderTicketInfoRepo,
		ticketGroupRepo,
		tagRepo,
		groupGalleryRepo,
		ticketDetailRepo,
		ticketGroupService,
		orderTicketLogRepo,
		generalRepo,
		customerService,
	)
	dashboardService := service.NewDashboardService( // ADD THESE LINES
		orderTicketGroupRepo,
		ticketGroupRepo,
		customerRepo,
		notificationRepo,
	)
	adminService := service.NewAdminServiceExtended(adminRepo, tokenRepo)
	tagService := service.NewTagService(tagRepo)
	bannerService := service.NewBannerService(bannerRepo)
	groupGalleryService := service.NewGroupGalleryService(groupGalleryRepo)
	pdfService := service.NewPDFService()
	notificationService := service.NewNotificationService(notificationRepo)
	generalService := service.NewGeneralService(generalRepo)

	// Initialize handlers
	adminHandler := handlers.NewAdminHandler(adminService, *notificationService)
	ticketGroupHandler := handlers.NewTicketGroupHandler(ticketGroupService, notificationService)
	authHandler := handlers.NewAuthHandler(authService, emailService, *customerService, *notificationService)
	customerHandler := handlers.NewCustomerHandler(*customerService, *notificationService)
	bannerHandler := handlers.NewBannerHandler(bannerService, *notificationService)
	groupGalleryHandler := handlers.NewGroupGalleryHandler(groupGalleryService)
	orderHandler := handlers.NewOrderHandler(orderService, *customerService, jwtService, paymentService, emailService, ticketGroupService, pdfService, *notificationService)
	paymentHandler := handlers.NewPaymentHandler(paymentService, emailService, ticketGroupService, pdfService, orderService, customerService, notificationService, generalRepo, cfg)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	pdfHandler := handlers.NewPDFHandler(pdfService)
	notificationHandler := handlers.NewNotificationHandler(*notificationService)
	tagHandler := handlers.NewTagHandler(tagService)
	generalHandler := handlers.NewGeneralHandler(generalService, *notificationService)

	// Create Fiber app with adapted error handler for slog
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler(slogger),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080,http://127.0.0.1:3000,http://localhost:3001,http://139.59.253.119:3002,https://etiket.johor.gov.my,http://stagingetiket.johor.gov.my:3002,https://stg-ticketcms.castis.io,http://stagingetiket.johor.gov.my:3001",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Static("/public", "./pkg/public")

	// Setup routes
	routes.SetupAdminRoutes(app, adminHandler, jwtService)
	routes.SetupTicketGroupRoutes(app, ticketGroupHandler, jwtService)
	routes.SetupAuthRoutes(app, authHandler, jwtService)
	routes.SetupOrderRoutes(app, orderHandler, jwtService)
	routes.SetupPaymentRoutes(app, paymentHandler, generalRepo, cfg)
	routes.SetupViewRoutes(app)
	routes.SetupCustomerRoutes(app, customerHandler, jwtService)
	routes.SetupBannerRoutes(app, bannerHandler, jwtService)
	routes.SetupGroupGalleryRoutes(app, groupGalleryHandler)
	routes.SetupTicketPDFRoutes(app, pdfHandler)
	routes.SetupDashboardRoutes(app, dashboardHandler, jwtService)
	routes.SetupNotificationRoutes(app, notificationHandler, jwtService)
	routes.SetupTagRoutes(app, tagHandler, jwtService)
	routes.SetupGeneralRoutes(app, generalHandler, jwtService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.CorePort)
	slogger.Info("Server starting",
		"address", addr,
		"version", "0.2.1")

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
