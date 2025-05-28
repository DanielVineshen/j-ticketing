// File: j-ticketing/internal/core/dto/order/order_response.go
package dto

import (
	ticketGroupDto "j-ticketing/internal/core/dto/ticket_group"
)

// OrderTicketGroupResponse represents the response structure for order ticket groups
type OrderTicketGroupResponse struct {
	OrderTicketGroups []OrderTicketGroupDTO `json:"orderTicketGroups"`
}

// OrderTicketGroupDTO represents the data transfer object for an order ticket group
type OrderTicketGroupDTO struct {
	OrderProfile  OrderProfileDTO                 `json:"orderProfile"`
	TicketProfile ticketGroupDto.TicketProfileDTO `json:"ticketProfile"`
}

// OrderProfileDTO represents the order profile data
type OrderProfileDTO struct {
	OrderTicketGroupId uint                 `json:"orderTicketGroupId"`
	TicketGroupId      uint                 `json:"ticketGroupId"`
	CustId             string               `json:"custId"`
	TransactionId      string               `json:"transactionId"`
	OrderNo            string               `json:"orderNo"`
	TransactionStatus  string               `json:"transactionStatus"`
	StatusMessage      string               `json:"statusMessage,omitempty"`
	TransactionDate    string               `json:"transactionDate"`
	BankCode           string               `json:"bankCode,omitempty"`
	BankName           string               `json:"bankName,omitempty"`
	MsgToken           string               `json:"msgToken"`
	BillId             string               `json:"billId"`
	ProductId          string               `json:"productId"`
	TotalAmount        float64              `json:"totalAmount"`
	BuyerName          string               `json:"buyerName"`
	BuyerEmail         string               `json:"buyerEmail"`
	ProductDesc        string               `json:"productDesc"`
	OrderTicketInfo    []OrderTicketInfoDTO `json:"orderTicketInfo"`
	OrderTicketLog     []OrderTicketLogDTO  `json:"orderTicketLog"`
	CreatedAt          string               `json:"createdAt"`
	UpdatedAt          string               `json:"updatedAt"`
}

// OrderTicketInfoDTO represents the order ticket information
type OrderTicketInfoDTO struct {
	OrderTicketInfoId  uint    `json:"orderTicketInfoId"`
	OrderTicketGroupId uint    `json:"orderTicketGroupId"`
	ItemId             string  `json:"itemId"`
	UnitPrice          float64 `json:"unitPrice"`
	ItemDesc1          string  `json:"itemDesc1"`
	ItemDesc2          string  `json:"itemDesc2"`
	PrintType          string  `json:"printType"`
	QuantityBought     int     `json:"quantityBought"`
	Twbid              string  `json:"twbid,omitempty"`
	EncryptedId        string  `json:"encryptedId"`
	AdmitDate          string  `json:"admitDate"`
	Variant            string  `json:"variant"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}

type OrderTicketLogDTO struct {
	OrderTicketLogId   uint   `json:"orderTicketLogId"`
	OrderTicketGroupId uint   `json:"orderTicketGroupId"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	Message            string `json:"message"`
	Date               string `json:"date"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}
