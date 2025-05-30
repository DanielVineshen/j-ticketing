package dto

type NotificationResponse struct {
	Notifications []NotificationDetails `json:"notifications"`
}

type NotificationDetails struct {
	NotificationID uint   `json:"notificationId"`
	PerformedBy    string `json:"performedBy"`
	AuthorityLevel string `json:"authorityLevel"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Date           string `json:"date"`
	IsRead         bool   `json:"isRead"`
	IsDeleted      bool   `json:"isDeleted"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}
