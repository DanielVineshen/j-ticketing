// File: internal/core/models/token.go
package models

import (
	"time"
)

// Token represents the token table
type Token struct {
	TokenId      uint      `gorm:"primaryKey;column:token_id;type:bigint unsigned AUTO_INCREMENT"`
	UserId       string    `gorm:"column:user_id;type:varchar(255)"`
	UserType     string    `gorm:"column:user_type;type:varchar(255)"`
	AccessToken  string    `gorm:"column:access_token;type:varchar(510)"`
	RefreshToken string    `gorm:"column:refresh_token;type:varchar(510)"`
	IpAddress    string    `gorm:"column:ip_address;type:varchar(255)"`
	UserAgent    string    `gorm:"column:user_agent;type:varchar(510)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

// TableName overrides the table name
func (Token) TableName() string {
	return "Token"
}
