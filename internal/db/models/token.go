// File: j-ticketing/internal/db/models/token.go
package models

import (
	"database/sql"
	"time"
)

// Token represents the token table
type Token struct {
	TokenId      uint           `gorm:"primaryKey;column:token_id;type:bigint unsigned AUTO_INCREMENT"`
	UserId       string         `gorm:"column:user_id;type:varchar(255);not null"`
	UserType     string         `gorm:"column:user_type;type:varchar(255);not null"`
	AccessToken  string         `gorm:"column:access_token;type:varchar(510);uniqueIndex;not null"`
	RefreshToken string         `gorm:"column:refresh_token;type:varchar(510);uniqueIndex;not null"`
	IpAddress    sql.NullString `gorm:"column:ip_address;type:varchar(255);null"`
	UserAgent    sql.NullString `gorm:"column:user_agent;type:varchar(510);null"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:datetime;not null"`
}

// TableName overrides the table name
func (Token) TableName() string {
	return "Token"
}
