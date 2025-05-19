package service

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
	"j-ticketing/pkg/email"
)

// PDFService handles PDF generation functionality
type PDFService struct {
	// Add any dependencies you might need here
}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{}
}

// GenerateTicketPDF generates a PDF ticket with QR codes
func (s *PDFService) GenerateTicketPDF(ticketGroupName string, tickets []email.TicketInfo) ([]byte, string, error) {
	// Create a new PDF with portrait orientation, mm unit, A4 format
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a new page
	pdf.AddPage()

	// Set default font
	pdf.SetFont("Arial", "B", 16)

	// Add title
	pdf.Cell(190, 10, ticketGroupName+" Tickets")
	pdf.Ln(20)

	// QR code settings
	qrSize := 40.0      // Size of QR code in mm
	marginX := 80.0     // Center margin in mm (A4 is 210mm wide)
	startY := 50.0      // Starting Y position in mm
	verticalGap := 65.0 // Gap between rows in mm
	itemsPerPage := 3   // Maximum items per page

	// Use RegisterImageOptionsReader to register in-memory images
	var qrOptions gofpdf.ImageOptions
	qrOptions.ImageType = "png"

	// Process tickets page by page
	for i := 0; i < len(tickets); i += itemsPerPage {
		// If not the first iteration, add a new page
		if i > 0 {
			pdf.AddPage()
		}

		// Process up to itemsPerPage items on the current page
		for j := 0; j < itemsPerPage && (i+j) < len(tickets); j++ {
			ticketIndex := i + j
			ticket := tickets[ticketIndex]

			// Calculate position (centered horizontally, stacked vertically)
			posY := startY + float64(j)*verticalGap

			// Generate QR code image
			qrCode, err := qr.Encode(ticket.Content, qr.M, qr.Auto)
			if err != nil {
				return nil, "", fmt.Errorf("failed to generate QR code: %w", err)
			}

			// Scale the QR code
			qrCode, err = barcode.Scale(qrCode, 256, 256)
			if err != nil {
				return nil, "", fmt.Errorf("failed to scale QR code: %w", err)
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
				return nil, "", fmt.Errorf("failed to encode QR code as PNG: %w", err)
			}

			// Register the in-memory image with gofpdf
			// The image ID must be unique for each image
			imageID := fmt.Sprintf("qr_%d", ticketIndex)
			pdf.RegisterImageOptionsReader(imageID, qrOptions, bytes.NewReader(qrBuffer.Bytes()))

			// Add QR code to PDF (centered horizontally)
			pdf.Image(imageID, marginX, posY, qrSize, 0, false, "", 0, "")

			// Add label under QR code
			pdf.SetFont("Arial", "", 10)
			pdf.SetY(posY + qrSize + 2)
			pdf.SetX(marginX)
			pdf.CellFormat(qrSize, 10, ticket.Label, "", 0, "C", false, 0, "")
		}
	}

	// Generate the PDF content
	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Generate a filename with the current timestamp
	filename := fmt.Sprintf("tickets_with_qrcodes_%s.pdf", time.Now().Format("20060102_150405"))

	// Return the PDF bytes and filename
	return buffer.Bytes(), filename, nil
}
