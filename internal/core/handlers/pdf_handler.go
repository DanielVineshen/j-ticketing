// File: j-ticketing/internal/core/handlers/pdf_handler.go
package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/email"
	"time"
)

// PDFHandler handles generating a PDF ticket
type PDFHandler struct {
	pdfService *service.PDFService
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(pdfService *service.PDFService) *PDFHandler {
	return &PDFHandler{
		pdfService: pdfService,
	}
}

// GenerateTicketPDF generates a PDF ticket with QR codes arranged horizontally
func (h *PDFHandler) GenerateTicketPDF(c *fiber.Ctx) error {
	orderOverview := email.OrderOverview{
		TicketGroup:  "Zoo Johor",
		FullName:     "Ahmad bin Abdullah",
		PurchaseDate: "2025-01-15 14:30:00",
		EntryDate:    "2025-01-20",
		Quantity:     5,
		OrderNumber:  "ORD-20250115143000-1234",
		Total:        125.00,
	}

	// Create OrderInfo array dummy data
	orderItems := []email.OrderInfo{
		{
			Description: "Dewasa Malaysia",
			Quantity:    2,
			Price:       30.00,
			EntryDate:   "2025-01-20",
		},
		{
			Description: "Kanak-kanak Malaysia (3-12 tahun)",
			Quantity:    2,
			Price:       15.00,
			EntryDate:   "2025-01-20",
		},
		{
			Description: "Warga Emas Malaysia (60+)",
			Quantity:    1,
			Price:       35.00,
			EntryDate:   "2025-01-20",
		},
	}

	// Create TicketInfo array dummy data
	ticketInfos := []email.TicketInfo{
		{
			Label:   "Dewasa Malaysia - Tiket 1",
			Content: "ZJ2025012000001",
		},
		{
			Label:   "Dewasa Malaysia - Tiket 2",
			Content: "ZJ2025012000002",
		},
		{
			Label:   "Kanak-kanak Malaysia - Tiket 1",
			Content: "ZJ2025012000003",
		},
		{
			Label:   "Kanak-kanak Malaysia - Tiket 2",
			Content: "ZJ2025012000004",
		},
		{
			Label:   "Warga Emas Malaysia - Tiket 1",
			Content: "ZJ2025012000005",
		},
	}

	bytes, _, _ := h.pdfService.GenerateTicketPDF(orderOverview, orderItems, ticketInfos, "en")

	// Generate a filename with the current timestamp
	filename := fmt.Sprintf("tickets_with_qrcodes_%s.pdf", time.Now().Format("20060102_150405"))

	// Set appropriate headers for PDF download
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename="+filename)

	// Send the PDF content
	return c.Send(bytes)
}
