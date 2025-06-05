// File: j-ticketing/pkg/config/config.go
package config

import (
	"github.com/joho/godotenv"
	logger "log/slog"
	"os"
	"strconv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// Email configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	UseSSL   bool
	// OAuth2 configuration
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// Zoo API configuration
type ZooAPIConfig struct {
	ZooBaseURL string
	Username   string
	Password   string
}

type Config struct {
	DB     DBConfig
	Server struct {
		CorePort        string
		SchedulerPort   string
		FrontendBaseUrl string
	}
	// Migration section to control migration behavior
	Migration struct {
		AutoMigrate bool
		Path        string
	}
	// JWT configuration
	JWT struct {
		SecretKey       string
		AccessTokenTTL  int64
		RefreshTokenTTL int64
	}
	// Email configuration
	Email EmailConfig
	// Zoo API configuration
	ZooAPI ZooAPIConfig
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logger.Info("Warning: .env file not found")
	}

	config := &Config{}

	// Database config
	config.DB.Host = getEnv("DB_HOST", "localhost")
	config.DB.Port = getEnv("DB_PORT", "3306")
	config.DB.User = getEnv("DB_USER", "root")
	config.DB.Password = getEnv("DB_PASSWORD", "")
	config.DB.Name = getEnv("DB_NAME", "ticketing")

	// Server config
	config.Server.CorePort = getEnv("SERVER_CORE_PORT", "8080")
	config.Server.SchedulerPort = getEnv("SERVER_SCHEDULER_PORT", "8081")
	config.Server.FrontendBaseUrl = getEnv("FRONTEND_BASE_URL", "http://localhost:3000")

	// Migration config
	config.Migration.AutoMigrate = getEnvBool("AUTO_MIGRATE", false)
	config.Migration.Path = getEnv("MIGRATION_PATH", "migrations")

	// JWT config
	config.JWT.SecretKey = getEnv("JWT_SECRET_KEY", "your-default-secret-key")
	config.JWT.AccessTokenTTL = getEnvInt64("JWT_ACCESS_TOKEN_TTL", 15)     // 15 minutes
	config.JWT.RefreshTokenTTL = getEnvInt64("JWT_REFRESH_TOKEN_TTL", 24*7) // 7 days

	// Email config
	//config.Email.Host = getEnv("EMAIL_HOST", "smtp.gmail.com")
	//config.Email.Port = getEnv("EMAIL_PORT", "587") // Default to 587 which works well with OAuth2
	//config.Email.Username = getEnv("EMAIL_USERNAME", "etiket@johor.gov.my")
	//config.Email.Password = getEnv("EMAIL_PASSWORD", "")
	//config.Email.From = getEnv("EMAIL_FROM", "etiket@johor.gov.my")

	// Automatically determine SSL usage based on port
	//port := config.Email.Port
	//if port == "465" {
	//	config.Email.UseSSL = true
	//	logger.Info("Using SSL mode for email (port 465)")
	//} else {
	//	config.Email.UseSSL = getEnvBool("EMAIL_USE_SSL", false)
	//	if port == "587" && !config.Email.UseSSL {
	//		logger.Info("Using STARTTLS mode for email (port 587)")
	//	} else if port == "587" && config.Email.UseSSL {
	//		logger.Info("Warning: Port 587 typically uses STARTTLS, not SSL. Consider setting EMAIL_USE_SSL=false")
	//	}
	//}

	// OAuth2 configuration
	//clientID := getEnv("CLIENT_ID", "")
	//// Check if client ID has a URL prefix and remove it
	//if strings.HasPrefix(clientID, "http://") || strings.HasPrefix(clientID, "https://") {
	//	logger.Info("Warning: CLIENT_ID contains a URL prefix. Removing prefix for OAuth2 authentication.")
	//	// Remove http:// or https:// prefix
	//	clientID = strings.TrimPrefix(strings.TrimPrefix(clientID, "http://"), "https://")
	//}
	//config.Email.ClientID = clientID
	//config.Email.ClientSecret = getEnv("CLIENT_SECRET", "")
	//config.Email.RefreshToken = getEnv("REFRESH_TOKEN", "")

	// Zoo API config
	//zooBaseURL := getEnv("ZOO_API_BASE_URL", "https://eglobal2.ddns.net/johorzooapi")
	//if !strings.HasPrefix(zooBaseURL, "http://") && !strings.HasPrefix(zooBaseURL, "https://") {
	//	zooBaseURL = "https://" + zooBaseURL
	//}
	//config.ZooAPI.ZooBaseURL = zooBaseURL
	//config.ZooAPI.Username = getEnv("ZOO_API_USERNAME", "")
	//config.ZooAPI.Password = getEnv("ZOO_API_PASSWORD", "")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}
