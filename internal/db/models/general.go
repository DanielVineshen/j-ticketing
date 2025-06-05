// File: j-ticketing/internal/db/models/general.go
package models

import (
	"time"
)

// General represents the general table
type General struct {
	GeneralId       uint   `gorm:"primaryKey;column:general_id;type:bigint unsigned AUTO_INCREMENT"`
	ContactEmail    string `gorm:"column:contact_email;type:varchar(255);not null"`
	TimeZone        string `gorm:"column:time_zone;type:varchar(255);not null"`
	EnvKey          string `gorm:"column:env_key;type:varchar(255);not null"`
	EnvValue        string `gorm:"column:env_value;type:varchar(255);not null"`
	AttachmentName  string `gorm:"column:attachment_name;type:varchar(255);not null"`
	AttachmentPath  string `gorm:"column:attachment_path;type:varchar(255);not null"`
	AttachmentSize  int64  `gorm:"column:attachment_size;type:bigint;not null"`
	ContentType     string `gorm:"column:content_type;type:varchar(255);not null"`
	UniqueExtension string `gorm:"column:unique_extension;type:varchar(255);not null"`

	// Content sections
	PrivacyPolicyContentEn         string `gorm:"column:privacy_policy_content_en;type:text;not null"`
	PrivacyPolicyContentBm         string `gorm:"column:privacy_policy_content_bm;type:text;not null"`
	PrivacyPolicyContentCn         string `gorm:"column:privacy_policy_content_cn;type:text;not null"`
	PrivacyPolicyLastUpdatedDate   string `gorm:"column:privacy_policy_last_updated_date;type:varchar(255);not null"`
	TermsOfPurchaseContentEn       string `gorm:"column:terms_of_purchase_content_en;type:text;not null"`
	TermsOfPurchaseContentBm       string `gorm:"column:terms_of_purchase_content_bm;type:text;not null"`
	TermsOfPurchaseContentCn       string `gorm:"column:terms_of_purchase_content_cn;type:text;not null"`
	TermsOfPurchaseLastUpdatedDate string `gorm:"column:terms_of_purchase_last_updated_date;type:varchar(255);not null"`
	TermsOfServiceContentEn        string `gorm:"column:terms_of_service_content_en;type:text;not null"`
	TermsOfServiceContentBm        string `gorm:"column:terms_of_service_content_bm;type:text;not null"`
	TermsOfServiceContentCn        string `gorm:"column:terms_of_service_content_cn;type:text;not null"`
	TermsOfServiceLastUpdatedDate  string `gorm:"column:terms_of_service_last_updated_date;type:varchar(255);not null"`
	FaqContentEn                   string `gorm:"column:faq_content_en;type:text;not null"`
	FaqContentBm                   string `gorm:"column:faq_content_bm;type:text;not null"`
	FaqContentCn                   string `gorm:"column:faq_content_cn;type:text;not null"`
	FaqLastUpdatedDate             string `gorm:"column:faq_last_updated_date;type:varchar(255);not null"`
	ContactUsContentEn             string `gorm:"column:contact_us_content_en;type:text;not null"`
	ContactUsContentBm             string `gorm:"column:contact_us_content_bm;type:text;not null"`
	ContactUsContentCn             string `gorm:"column:contact_us_content_cn;type:text;not null"`
	ContactUsLastUpdatedDate       string `gorm:"column:contact_us_last_updated_date;type:varchar(255);not null"`
	RefundPolicyContentEn          string `gorm:"column:refund_policy_content_en;type:text;not null"`
	RefundPolicyContentBm          string `gorm:"column:refund_policy_content_bm;type:text;not null"`
	RefundPolicyContentCn          string `gorm:"column:refund_policy_content_cn;type:text;not null"`
	RefundPolicyLastUpdatedDate    string `gorm:"column:refund_policy_last_updated_date;type:varchar(255);not null"`

	// Zoo API Configuration
	ZooApiBaseUrl    string `gorm:"column:zoo_api_base_url;type:varchar(255);not null"`
	ZooQrEndpoint    string `gorm:"column:zoo_qr_endpoint;type:varchar(255);not null"`
	ZooTokenEndpoint string `gorm:"column:zoo_token_endpoint;type:varchar(255);not null"`
	ZooApiUsername   string `gorm:"column:zoo_api_username;type:varchar(255);not null"`
	ZooApiPassword   string `gorm:"column:zoo_api_password;type:varchar(255);not null"`

	// JohorPay Configuration
	JpGatewayUrl       string `gorm:"column:jp_gateway_url;type:varchar(255);not null"`
	JpPaymentEndpoint  string `gorm:"column:jp_payment_endpoint;type:varchar(255);not null"`
	JpRedflowEndpoint  string `gorm:"column:jp_redflow_endpoint;type:varchar(255);not null"`
	JpBankListEndpoint string `gorm:"column:jp_bank_list_endpoint;type:varchar(255);not null"`
	JpApiKey           string `gorm:"column:jp_api_key;type:varchar(255);not null"`
	JpAgToken          string `gorm:"column:jp_ag_token;type:varchar(255);not null"`

	// Email Configuration
	EmailHost         string `gorm:"column:email_host;type:varchar(255);not null"`
	EmailPort         int    `gorm:"column:email_port;type:int;not null"`
	EmailUsername     string `gorm:"column:email_username;type:varchar(255);not null"`
	EmailPassword     string `gorm:"column:email_password;type:varchar(255);not null"`
	EmailFrom         string `gorm:"column:email_from;type:varchar(255);not null"`
	EmailUseSsl       bool   `gorm:"column:email_use_ssl;type:boolean;not null"`
	EmailClientId     string `gorm:"column:email_client_id;type:varchar(510);not null"`
	EmailClientSecret string `gorm:"column:email_client_secret;type:varchar(255);not null"`
	EmailRefreshToken string `gorm:"column:email_refresh_token;type:varchar(255);not null"`

	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName overrides the table name
func (General) TableName() string {
	return "General"
}
