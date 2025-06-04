// File: j-ticketing/internal/core/dto/general/general_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

type UpdateGeneralRequest struct {
	ContactEmail string `json:"contactEmail" validate:"required,email,max=255"`
	TimeZone     string `json:"timeZone" validate:"required,max=255"`
}

func (r *UpdateGeneralRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdatePrivacyPolicyRequest for updating privacy policy content
type UpdatePrivacyPolicyRequest struct {
	PrivacyPolicyContentEn       string `json:"privacyPolicyContentEn" validate:"required"`
	PrivacyPolicyContentBm       string `json:"privacyPolicyContentBm" validate:"required"`
	PrivacyPolicyContentCn       string `json:"privacyPolicyContentCn" validate:"required"`
	PrivacyPolicyLastUpdatedDate string `json:"privacyPolicyLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdatePrivacyPolicyRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateTermsOfPurchaseRequest for updating terms of purchase content
type UpdateTermsOfPurchaseRequest struct {
	TermsOfPurchaseContentEn       string `json:"termsOfPurchaseContentEn" validate:"required"`
	TermsOfPurchaseContentBm       string `json:"termsOfPurchaseContentBm" validate:"required"`
	TermsOfPurchaseContentCn       string `json:"termsOfPurchaseContentCn" validate:"required"`
	TermsOfPurchaseLastUpdatedDate string `json:"termsOfPurchaseLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdateTermsOfPurchaseRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateTermsOfServiceRequest for updating terms of service content
type UpdateTermsOfServiceRequest struct {
	TermsOfServiceContentEn       string `json:"termsOfServiceContentEn" validate:"required"`
	TermsOfServiceContentBm       string `json:"termsOfServiceContentBm" validate:"required"`
	TermsOfServiceContentCn       string `json:"termsOfServiceContentCn" validate:"required"`
	TermsOfServiceLastUpdatedDate string `json:"termsOfServiceLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdateTermsOfServiceRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateFaqRequest for updating FAQ content
type UpdateFaqRequest struct {
	FaqContentEn       string `json:"faqContentEn" validate:"required"`
	FaqContentBm       string `json:"faqContentBm" validate:"required"`
	FaqContentCn       string `json:"faqContentCn" validate:"required"`
	FaqLastUpdatedDate string `json:"faqLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdateFaqRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateContactUsRequest for updating contact us content
type UpdateContactUsRequest struct {
	ContactUsContentEn       string `json:"contactUsContentEn" validate:"required"`
	ContactUsContentBm       string `json:"contactUsContentBm" validate:"required"`
	ContactUsContentCn       string `json:"contactUsContentCn" validate:"required"`
	ContactUsLastUpdatedDate string `json:"contactUsLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdateContactUsRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateRefundPolicyRequest for updating refund policy content
type UpdateRefundPolicyRequest struct {
	RefundPolicyContentEn       string `json:"refundPolicyContentEn" validate:"required"`
	RefundPolicyContentBm       string `json:"refundPolicyContentBm" validate:"required"`
	RefundPolicyContentCn       string `json:"refundPolicyContentCn" validate:"required"`
	RefundPolicyLastUpdatedDate string `json:"refundPolicyLastUpdatedDate" validate:"required,max=255"`
}

func (r *UpdateRefundPolicyRequest) Validate() error {
	return validation.ValidateStruct(r)
}
