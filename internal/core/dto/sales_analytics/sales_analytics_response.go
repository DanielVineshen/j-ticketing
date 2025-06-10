// File: j-ticketing/internal/core/dto/sales_analytics/sales_analytics_response.go
package dto

// TotalRevenueResponse represents the response for total revenue API
type TotalRevenueResponse struct {
	SumTotalRevenue float64        `json:"sumTotalRevenue"`
	RevenueTrend    []RevenueTrend `json:"revenueTrend"`
}

type RevenueTrend struct {
	Date         string  `json:"date"`
	TotalRevenue float64 `json:"totalRevenue"`
}

// TotalOrdersResponse represents the response for total orders API
type TotalOrdersResponse struct {
	SumTotalOrders int          `json:"sumTotalOrders"`
	OrderTrend     []OrderTrend `json:"orderTrend"`
}

type OrderTrend struct {
	Date        string `json:"date"`
	TotalOrders int    `json:"totalOrders"`
}

// AvgOrderValueResponse represents the response for average order value API
type AvgOrderValueResponse struct {
	TotalAvgOrderValue float64              `json:"totalAvgOrderValue"`
	AvgOrderValueTrend []AvgOrderValueTrend `json:"avgOrderValueTrend"`
}

type AvgOrderValueTrend struct {
	Date          string  `json:"date"`
	AvgOrderValue float64 `json:"avgOrdervalue"`
}

// TopSalesProductResponse represents the response for top sales product API
type TopSalesProductResponse struct {
	TopSaleProduct      TopSaleProduct        `json:"topSaleProduct"`
	TopSaleProductTrend []TopSaleProductTrend `json:"topSaleProductTrend"`
}

type TopSaleProduct struct {
	TicketGroupName string `json:"ticketGroupName"`
	SumTotalOrders  int    `json:"sumTotalOrders"`
}

type TopSaleProductTrend struct {
	TicketGroupName string  `json:"ticketGroupName"`
	TotalQuantity   int     `json:"totalQuantity"`
	TotalRevenue    float64 `json:"totalRevenue"`
	TotalOrders     int     `json:"totalOrders"`
}
