package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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
}

// Zoo API configuration
type ZooAPIConfig struct {
	BaseURL  string
	Username string
	Password string
}

type Config struct {
	DB     DBConfig
	Server struct {
		CorePort      string
		SchedulerPort string
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
		fmt.Println("Warning: .env file not found")
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

	// Migration config
	config.Migration.AutoMigrate = getEnvBool("AUTO_MIGRATE", false)
	config.Migration.Path = getEnv("MIGRATION_PATH", "migrations")

	// JWT config
	config.JWT.SecretKey = getEnv("JWT_SECRET_KEY", "your-default-secret-key")
	config.JWT.AccessTokenTTL = getEnvInt64("JWT_ACCESS_TOKEN_TTL", 15)     // 15 minutes
	config.JWT.RefreshTokenTTL = getEnvInt64("JWT_REFRESH_TOKEN_TTL", 24*7) // 7 days

	// Email config
	config.Email.Host = getEnv("EMAIL_HOST", "smtp.gmail.com")
	config.Email.Port = getEnv("EMAIL_PORT", "465") // Default to SSL port
	config.Email.Username = getEnv("EMAIL_USERNAME", "etiket@johor.gov.my")
	config.Email.Password = getEnv("EMAIL_PASSWORD", "")
	config.Email.From = getEnv("EMAIL_FROM", "etiket@johor.gov.my")
	config.Email.UseSSL = getEnvBool("EMAIL_USE_SSL", true) // Default to SSL

	// Zoo API config
	baseURL := getEnv("ZOO_API_BASE_URL", "https://eglobal2.ddns.net/johorzooapi")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	config.ZooAPI.BaseURL = baseURL
	config.ZooAPI.Username = getEnv("ZOO_API_USERNAME", "")
	config.ZooAPI.Password = getEnv("ZOO_API_PASSWORD", "")

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
