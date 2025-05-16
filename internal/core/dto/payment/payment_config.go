package payment

// PaymentConfig holds our payment gateway configuration
type PaymentConfig struct {
	GatewayURL      string
	APIKey          string
	BaseURL         string
	AGToken         string
	FrontendBaseURL string
}
