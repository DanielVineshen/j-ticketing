// File: j-ticketing/internal/db/models/order_ticket_group.go
package models

import (
	"database/sql"
	"time"
)

// OrderTicketGroup represents the order_ticket_group table
type OrderTicketGroup struct {
	OrderTicketGroupId uint           `gorm:"primaryKey;column:order_ticket_group_id;type:bigint unsigned AUTO_INCREMENT"`
	LangChosen         string         `gorm:"column:lang_chosen;type:varchar(255);default:'bm';not null"`
	TicketGroupId      uint           `gorm:"column:ticket_group_id;type:bigint unsigned;not null"`
	CustId             string         `gorm:"column:cust_id;type:varchar(255);not null"`
	TransactionId      string         `gorm:"column:transaction_id;type:varchar(255);not null"`
	OrderNo            string         `gorm:"column:order_no;type:varchar(255);not null"`
	TransactionStatus  string         `gorm:"column:transaction_status;type:varchar(255);not null"`
	BankCurrentStatus  string         `gorm:"column:bank_current_status;type:varchar(255);null"`
	StatusMessage      sql.NullString `gorm:"column:status_message;type:varchar(255);null"`
	TransactionDate    string         `gorm:"column:transaction_date;type:varchar(255);not null"`
	BankCode           sql.NullString `gorm:"column:bank_code;type:varchar(255);null"`
	BankName           sql.NullString `gorm:"column:bank_name;type:varchar(255);null"`
	MsgToken           string         `gorm:"column:msg_token;type:varchar(255);not null"`
	BillId             string         `gorm:"column:bill_id;type:varchar(255);not null"`
	ProductId          string         `gorm:"column:product_id;type:varchar(255);not null"`
	TotalAmount        float64        `gorm:"column:total_amount;type:decimal(10,2);not null"`
	BuyerName          string         `gorm:"column:buyer_name;type:varchar(255);not null"`
	BuyerEmail         string         `gorm:"column:buyer_email;type:varchar(255);not null"`
	ProductDesc        string         `gorm:"column:product_desc;type:varchar(255);not null"`
	IsEmailSent        bool           `gorm:"column:is_email_sent;type:boolean;default:false;not null"`
	AdmitDate          string         `gorm:"column:admit_date;type:varchar(255);not null"`
	CreatedAt          time.Time      `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships with constraint:false to prevent auto-generation of constraints
	TicketGroup      TicketGroup       `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
	Customer         Customer          `gorm:"foreignKey:CustId;references:CustId;constraint:false"`
	OrderTicketInfos []OrderTicketInfo `gorm:"foreignKey:OrderTicketGroupId"`
	OrderTicketLogs  []OrderTicketLog  `gorm:"foreignKey:OrderTicketGroupId"`
}

// TableName overrides the table name
func (OrderTicketGroup) TableName() string {
	return "Order_Ticket_Group"
}
