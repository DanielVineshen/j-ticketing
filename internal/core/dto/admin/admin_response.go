// File: j-ticketing/internal/core/dto/admin/admin_response.go
package dto

type AdminProfileResponse struct {
	Admin AdminProfile `json:"adminProfile"`
}

type AdminProfile struct {
	AdminID   int    `json:"adminId"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	ContactNo string `json:"contactNo"`
	Role      string `json:"role"`
}

type AllAdminResponse struct {
	Admins []AdminManagement `json:"admins"`
}

type AdminManagement struct {
	AdminID    int    `json:"adminId"`
	Username   string `json:"username"`
	FullName   string `json:"fullName"`
	Email      string `json:"email"`
	ContactNo  string `json:"contactNo"`
	Role       string `json:"role"`
	IsDisabled bool   `json:"isDisabled"`
}
