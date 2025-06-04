// File: j-ticketing/internal/core/dto/general/general_response.go
package dto

type GeneralResponse struct {
	GeneralId                      uint   `json:"generalId"`
	ContactEmail                   string `json:"contactEmail"`
	TimeZone                       string `json:"timeZone"`
	EnvKey                         string `json:"envKey"`
	EnvValue                       string `json:"envValue"`
	AttachmentName                 string `json:"attachmentName"`
	ContentType                    string `json:"contentType"`
	UniqueExtension                string `json:"uniqueExtension"`
	PrivacyPolicyContentEn         string `json:"privacyPolicyContentEn"`
	PrivacyPolicyContentBm         string `json:"privacyPolicyContentBm"`
	PrivacyPolicyContentCn         string `json:"privacyPolicyContentCn"`
	PrivacyPolicyLastUpdatedDate   string `json:"privacyPolicyLastUpdatedDate"`
	TermsOfPurchaseContentEn       string `json:"termsOfPurchaseContentEn"`
	TermsOfPurchaseContentBm       string `json:"termsOfPurchaseContentBm"`
	TermsOfPurchaseContentCn       string `json:"termsOfPurchaseContentCn"`
	TermsOfPurchaseLastUpdatedDate string `json:"termsOfPurchaseLastUpdatedDate"`
	TermsOfServiceContentEn        string `json:"termsOfServiceContentEn"`
	TermsOfServiceContentBm        string `json:"termsOfServiceContentBm"`
	TermsOfServiceContentCn        string `json:"termsOfServiceContentCn"`
	TermsOfServiceLastUpdatedDate  string `json:"termsOfServiceLastUpdatedDate"`
	FaqContentEn                   string `json:"faqContentEn"`
	FaqContentBm                   string `json:"faqContentBm"`
	FaqContentCn                   string `json:"faqContentCn"`
	FaqLastUpdatedDate             string `json:"faqLastUpdatedDate"`
	ContactUsContentEn             string `json:"contactUsContentEn"`
	ContactUsContentBm             string `json:"contactUsContentBm"`
	ContactUsContentCn             string `json:"contactUsContentCn"`
	ContactUsLastUpdatedDate       string `json:"contactUsLastUpdatedDate"`
	RefundPolicyContentEn          string `json:"refundPolicyContentEn"`
	RefundPolicyContentBm          string `json:"refundPolicyContentBm"`
	RefundPolicyContentCn          string `json:"refundPolicyContentCn"`
	RefundPolicyLastUpdatedDate    string `json:"refundPolicyLastUpdatedDate"`
	CreatedAt                      string `json:"createdAt"` // yyyy-MM-dd HH:mm:ss format (Malaysia time)
	UpdatedAt                      string `json:"updatedAt"` // yyyy-MM-dd HH:mm:ss format (Malaysia time)
}
