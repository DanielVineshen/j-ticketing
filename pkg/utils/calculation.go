// File: j-ticketing/pkg/utils/calculator.go
package utils

import (
	"j-ticketing/pkg/email"
)

func CalculateOrderTotal(orders []email.OrderInfo) float64 {
	var total = 0.0

	for _, order := range orders {
		quantity := order.Quantity
		price := order.Price
		total += float64(quantity) * price
	}

	return total
}
