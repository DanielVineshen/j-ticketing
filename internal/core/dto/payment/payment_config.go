// File: j-ticketing/internal/core/dto/payment/payment_config.go
package payment

// PaymentConfig holds our payment gateway configuration
type PaymentConfig struct {
	GatewayURL      string
	APIKey          string
	BaseURL         string
	AGToken         string
	FrontendBaseURL string
}
