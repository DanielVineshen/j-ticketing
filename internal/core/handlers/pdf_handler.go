package handlers

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"image"
	"image/png"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
)

// PDFHandler handles generating a PDF ticket
type PDFHandler struct {
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler() *PDFHandler {
	return &PDFHandler{}
}

// GenerateTicketPDF generates a PDF ticket with QR codes arranged horizontally
func (h *PDFHandler) GenerateTicketPDF(c *fiber.Ctx) error {
	ticketGroupName := "Zoo Johor"

	// QR code content
	tickets := []struct {
		Content string
		Label   string
	}{
		{"TICKET-ADULT-001", "Adult Ticket #001"},
		{"TICKET-ADULT-002", "Adult Ticket #002"},
		{"TICKET-CHILD-001", "Child Ticket #001"},
		{"TICKET-CHILD-002", "Child Ticket #002"},
		{"TICKET-SENIOR-001", "Senior Ticket #001"},
		{"TICKET-SENIOR-002", "Senior Ticket #002"},
		{"TICKET-SENIOR-002", "Senior Ticket #002"},
	}

	// Create a new PDF with portrait orientation, mm unit, A4 format
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Set default font
	pdf.SetFont("Arial", "B", 16)

	// QR code settings
	qrSize := 25.0        // Size of QR code in mm
	startX := 15.0        // Starting X position in mm (left margin)
	horizontalGap := 35.0 // Gap between QR codes in mm
	startY := 30.0        // Starting Y position in mm
	itemsPerRow := 5      // Maximum items per row
	itemsPerPage := 10    // Maximum items per page
	labelOffset := 2.0    // Space between QR code and its label in mm

	// Use RegisterImageOptionsReader to register in-memory images
	var qrOptions gofpdf.ImageOptions
	qrOptions.ImageType = "png"

	// Process tickets page by page
	for i := 0; i < len(tickets); i += itemsPerPage {
		// Add a new page
		pdf.AddPage()

		// Add title
		pdf.SetY(10)
		pdf.Cell(190, 10, ticketGroupName+" Tickets")

		// Calculate how many items will be on this page
		itemsOnPage := itemsPerPage
		if i+itemsPerPage > len(tickets) {
			itemsOnPage = len(tickets) - i
		}

		// Process up to itemsPerPage items on the current page
		for j := 0; j < itemsOnPage; j++ {
			ticketIndex := i + j
			ticket := tickets[ticketIndex]

			// Calculate row and column indices
			row := j / itemsPerRow
			col := j % itemsPerRow

			// Calculate position (arranged horizontally, then wrap to next row)
			posX := startX + float64(col)*horizontalGap
			posY := startY + float64(row)*50.0 // 90mm vertical spacing between rows

			// Generate QR code image
			qrCode, err := qr.Encode(ticket.Content, qr.M, qr.Auto)
			if err != nil {
				return fmt.Errorf("failed to generate QR code: %w", err)
			}

			// Scale the QR code
			qrCode, err = barcode.Scale(qrCode, 256, 256)
			if err != nil {
				return fmt.Errorf("failed to scale QR code: %w", err)
			}

			// Convert to RGBA to ensure compatibility
			rgba := image.NewRGBA(qrCode.Bounds())
			for y := 0; y < qrCode.Bounds().Dy(); y++ {
				for x := 0; x < qrCode.Bounds().Dx(); x++ {
					rgba.Set(x, y, qrCode.At(x, y))
				}
			}

			// Create an in-memory buffer for the PNG
			var qrBuffer bytes.Buffer
			err = png.Encode(&qrBuffer, rgba)
			if err != nil {
				return fmt.Errorf("failed to encode QR code as PNG: %w", err)
			}

			// Register the in-memory image with gofpdf
			// The image ID must be unique for each image
			imageID := fmt.Sprintf("qr_%d", ticketIndex)
			pdf.RegisterImageOptionsReader(imageID, qrOptions, bytes.NewReader(qrBuffer.Bytes()))

			// Add QR code to PDF at calculated position
			pdf.Image(imageID, posX, posY, qrSize, 0, false, "", 0, "")

			// Add label under QR code
			pdf.SetFont("Arial", "", 10)
			pdf.SetY(posY + qrSize + labelOffset)
			pdf.SetX(posX)
			pdf.CellFormat(qrSize, 10, ticket.Label, "", 0, "C", false, 0, "")

			// Reset to bold font for the next ticket
			pdf.SetFont("Arial", "B", 16)
		}
	}

	// Generate the PDF content
	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Generate a filename with the current timestamp
	filename := fmt.Sprintf("tickets_with_qrcodes_%s.pdf", time.Now().Format("20060102_150405"))

	// Set appropriate headers for PDF download
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename="+filename)

	// Send the PDF content
	return c.Send(buffer.Bytes())
}
