package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
)

// PDFHandler handles generating a PDF ticket
type PDFHandler struct{}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler() *PDFHandler {
	return &PDFHandler{}
}

// GenerateTicketPDF generates a PDF ticket with QR codes
func (h *PDFHandler) GenerateTicketPDF(c *fiber.Ctx) error {
	// Create a new PDF with portrait orientation, mm unit, A4 format
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a new page
	pdf.AddPage()

	// Set default font
	pdf.SetFont("Arial", "B", 16)

	// Add title
	pdf.Cell(190, 10, "Zoo Johor Tickets")
	pdf.Ln(20)

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
	}

	// Create temp directory if it doesn't exist
	tmpDir, err := ioutil.TempDir("", "qrcodes")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create temp directory: " + err.Error(),
		})
	}
	defer os.RemoveAll(tmpDir)

	// QR code settings
	qrSize := 40.0      // Size of QR code in mm
	marginX := 80.0     // Center margin in mm (A4 is 210mm wide)
	startY := 50.0      // Starting Y position in mm
	verticalGap := 65.0 // Gap between rows in mm
	itemsPerPage := 3   // Maximum items per page

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
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate QR code: " + err.Error(),
				})
			}

			// Scale the QR code
			qrCode, err = barcode.Scale(qrCode, 256, 256)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to scale QR code: " + err.Error(),
				})
			}

			// Convert to RGBA to ensure compatibility
			rgba := image.NewRGBA(qrCode.Bounds())
			for y := 0; y < qrCode.Bounds().Dy(); y++ {
				for x := 0; x < qrCode.Bounds().Dx(); x++ {
					rgba.Set(x, y, qrCode.At(x, y))
				}
			}

			// Create a file to save the QR code
			qrCodeFile := filepath.Join(tmpDir, fmt.Sprintf("qr_code_%d.png", ticketIndex))
			file, err := os.Create(qrCodeFile)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create QR code file: " + err.Error(),
				})
			}

			// Encode the QR code as PNG
			err = png.Encode(file, rgba)
			file.Close()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to encode QR code: " + err.Error(),
				})
			}

			// Add QR code to PDF (centered horizontally)
			pdf.Image(qrCodeFile, marginX, posY, qrSize, 0, false, "", 0, "")

			// Add label under QR code
			pdf.SetFont("Arial", "", 10)
			pdf.SetY(posY + qrSize + 2)
			pdf.SetX(marginX)
			pdf.CellFormat(qrSize, 10, ticket.Label, "", 0, "C", false, 0, "")
		}
	}

	// Generate the PDF content
	var buffer bytes.Buffer
	err = pdf.Output(&buffer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate PDF: " + err.Error(),
		})
	}

	// Set appropriate headers for PDF download
	c.Set("Content-Type", "application/pdf")

	// Generate a filename with the current timestamp
	filename := fmt.Sprintf("tickets_with_qrcodes_%s.pdf", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", "attachment; filename="+filename)

	// Send the PDF content
	return c.Send(buffer.Bytes())
}
