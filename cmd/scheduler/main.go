// File: j-ticketing/cmd/scheduler/main.go
package main

import (
	"fmt"
	service "j-ticketing/internal/core/services"
	"j-ticketing/internal/db"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/internal/scheduler/jobs"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/middleware"
	"log"
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

	// Initialize repositories
	ticketGroupRepo := repositories.NewTicketGroupRepository(database)
	tagRepo := repositories.NewTagRepository(database)
	groupGalleryRepo := repositories.NewGroupGalleryRepository(database)
	ticketDetailRepo := repositories.NewTicketDetailRepository(database)
	orderTicketGroupRepo := repositories.NewOrderTicketGroupRepository(database)
	orderTicketInfoRepo := repositories.NewOrderTicketInfoRepository(database)
	orderTicketLogRepo := repositories.NewOrderTicketLogRepository(database)
	customerRepo := repositories.NewCustomerRepository(database)
	customerLogRepo := repositories.NewCustomerLogRepository(database)
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
	pdfService := service.NewPDFService()

	emailProcessingService := jobs.NewEmailProcessingService(paymentService, orderTicketGroupRepo, orderTicketInfoRepo, emailService, ticketGroupService, pdfService, orderService)

	go runScheduler(emailProcessingService)

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

// runScheduler runs the scheduler in a loop
func runScheduler(orderService *jobs.EmailProcessingService) {
	log.Printf("INFO: Starting scheduler")

	for {
		// Log the start of the scheduler run
		startTime := time.Now()
		log.Printf("INFO: Running scheduled task: Processing orders")

		// Process orders - wait for completion
		count, err := orderService.ProcessPendingOrders()
		if err != nil {
			log.Printf("ERROR: Error processing orders: %v", err)
		} else {
			log.Printf("INFO: Order processing complete (processed_count=%d)", count)
		}

		// Calculate how long the processing took
		processingDuration := time.Since(startTime)
		nextRunAt := time.Now().Add(2 * time.Minute)

		// Log completion and wait time
		log.Printf("INFO: Scheduler run complete, waiting 2 minutes until next run (processing_duration=%v, next_run_at=%v)",
			processingDuration, nextRunAt.Format(time.RFC3339))

		// Wait exactly 2 minutes before the next run, regardless of how long processing took
		time.Sleep(2 * time.Minute)
	}
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
