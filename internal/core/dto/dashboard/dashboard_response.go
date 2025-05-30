// File: j-ticketing/internal/core/dto/dashboard/dashboard_response.go
package dto

type DashboardResponse struct {
	OrdersAnalysis   interface{} `json:"ordersAnalysis"`
	ProductAnalysis  interface{} `json:"productAnalysis"`
	CustomerAnalysis interface{} `json:"customerAnalysis"`
	Notifications    interface{} `json:"notifications"`
}
