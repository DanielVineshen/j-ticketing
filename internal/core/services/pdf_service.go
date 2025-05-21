package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/storage"
	"strconv"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
)

// PDFService handles PDF generation functionality
type PDFService struct {
	// Add any dependencies you might need here
}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{}
}

// Item represents an order item
type Item struct {
	Name        string
	Description string
	Price       float64
	Quantity    int
	Date        string
}

// Order represents order details
type Order struct {
	OrderID        string
	PlacedAt       string
	Total          float64
	Status         string
	Items          []Item
	DiscountAmount float64
	GST            bool
}

func addHeader(logoBase64 string, ticketGroupName string, addr1 string, addr2 string, generalLine string, pdf *gofpdf.Fpdf) {
	// Header measurements
	headerHeight := 50.0
	maxLogoWidth := 50.0  // Maximum width constraint
	maxLogoHeight := 45.0 // Maximum height constraint
	pageWidth := 210.0    // A4 width in mm

	logoBytes, err := base64.StdEncoding.DecodeString(logoBase64)
	if err == nil {
		// Create an image reader
		imgReader := bytes.NewReader(logoBytes)

		// Get the image config to determine dimensions
		img, _, err := image.DecodeConfig(imgReader)
		if err == nil {
			// Reset reader position
			imgReader.Seek(0, 0)

			// Calculate aspect ratio
			aspectRatio := float64(img.Width) / float64(img.Height)

			// Determine scaled dimensions while maintaining aspect ratio
			var logoWidth, logoHeight float64

			if aspectRatio > 1 {
				// Image is wider than tall
				logoWidth = maxLogoWidth
				logoHeight = logoWidth / aspectRatio

				// Check if height exceeds maximum
				if logoHeight > maxLogoHeight {
					logoHeight = maxLogoHeight
					logoWidth = logoHeight * aspectRatio
				}
			} else {
				// Image is taller than wide
				logoHeight = maxLogoHeight
				logoWidth = logoHeight * aspectRatio

				// Check if width exceeds maximum
				if logoWidth > maxLogoWidth {
					logoWidth = maxLogoWidth
					logoHeight = logoWidth / aspectRatio
				}
			}

			// Register and place the image with calculated dimensions
			var logoOptions gofpdf.ImageOptions
			logoOptions.ImageType = "png"
			pdf.RegisterImageOptionsReader("logo", logoOptions, imgReader)
			pdf.Image("logo", 20, 25, logoWidth, logoHeight, false, "", 0, "")
		} else {
			// Fallback to original code if we can't get image dimensions
			var logoOptions gofpdf.ImageOptions
			logoOptions.ImageType = "png"
			pdf.RegisterImageOptionsReader("logo", logoOptions, imgReader)
			pdf.Image("logo", 20, 25, maxLogoWidth, maxLogoHeight, false, "", 0, "")
		}
	}

	// Add the title box on the right side
	boxWidth := 115.0
	boxHeight := headerHeight
	boxX := pageWidth - boxWidth - 15 // 15mm margin from right
	boxY := 20.0                      // 10mm from top

	// Draw the rounded rectangle with orange background
	pdf.SetFillColor(213, 197, 138) // Khaki to match the header
	pdf.RoundedRect(boxX, boxY, boxWidth, boxHeight, 10, strconv.Itoa(10), "F")

	// Add text to the box
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 0, 0) // Black text

	// Title - use MultiCell for auto-wrapping long text
	titleY := boxY + 5 // Start a bit higher to accommodate potentially two lines
	pdf.SetXY(boxX+5, titleY)
	pdf.MultiCell(boxWidth-10, 8, strings.ToUpper(ticketGroupName), "", "C", false)

	// Get current Y position after title, to make sure address lines don't overlap
	currentY := pdf.GetY() + 3 // Add small gap after title

	// Address lines
	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(boxX, currentY)
	pdf.CellFormat(boxWidth, 8, addr1, "", 0, "C", false, 0, "")

	pdf.SetXY(boxX, currentY+8)
	pdf.CellFormat(boxWidth, 8, addr2, "", 0, "C", false, 0, "")

	// General line
	pdf.SetXY(boxX, currentY+16)
	pdf.CellFormat(boxWidth, 8, "Talian Umum: "+generalLine, "", 0, "C", false, 0, "")

	// Reset text color to black for the rest of the document
	pdf.SetTextColor(0, 0, 0)
}

// addParticipantInfo adds the participant information section
func addParticipantInfo(pdf *gofpdf.Fpdf, participantName string, purchaseDate string, entryDate string, orderNo string, totalTickets int) {
	startY := 85.0 // Start position after the header
	leftColX := 15.0
	midColX := 80.0
	rightColX := 150.0

	// Column titles - Left column
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(leftColX, startY)
	pdf.Cell(100, 5, "Peserta Utama")

	// Column titles - Middle column
	pdf.SetXY(midColX, startY)
	pdf.Cell(100, 5, "Tarikh Pembelian")

	// Column titles - Right column
	pdf.SetXY(rightColX, startY)
	pdf.Cell(100, 5, "Tarikh Masuk")

	// Values - Left column (Participant Name)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black for the values
	pdf.SetXY(leftColX, startY+7)
	pdf.Cell(100, 5, participantName)

	// Values - Middle column (Purchase Date)
	pdf.SetXY(midColX, startY+7)
	pdf.Cell(100, 5, purchaseDate)

	// Values - Right column (Entry Date)
	pdf.SetXY(rightColX, startY+7)
	pdf.Cell(100, 5, entryDate)

	// Total Tickets - Label
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(leftColX, startY+18)
	pdf.Cell(100, 5, "Jumlah Tiket")

	// Order No. - Label
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(midColX, startY+18)
	pdf.Cell(100, 5, "No. Pesanan")

	// Order No. - Value
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black for the values
	pdf.SetXY(midColX, startY+25)
	pdf.Cell(100, 5, orderNo)

	// Total Tickets - Value
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black for the values
	pdf.SetXY(leftColX, startY+25)
	pdf.Cell(100, 5, strconv.Itoa(totalTickets))
}

// addRedeemSection adds the redeem instructions section
func addRedeemSection(pdf *gofpdf.Fpdf) {
	startY := 130.0 // Start position after the participant info
	leftX := 15.0

	// Add section heading
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(leftX, startY)
	pdf.Cell(180, 5, "Tebus Unit Individu")

	// Add instruction text
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(leftX, startY+7)
	pdf.Cell(180, 5, "Imbas kod QR di bawah untuk menebus unit anda secara individu.")
}

// addTermsAndConditionsPage adds a new page with an enhanced Terms and Conditions content
func addTermsAndConditionsPage(ticketGroupName string, contactNo string, email string, pdf *gofpdf.Fpdf) {
	pdf.AddPage()

	// Set margins for the terms page
	leftMargin := 15.0
	topMargin := 15.0
	pdf.SetMargins(leftMargin, topMargin, 15.0)

	// Page width calculation for design elements
	pageWidth := 210.0                            // A4 width in mm
	contentWidth := pageWidth - leftMargin - 15.0 // Left and right margins

	// Add a stylish header bar for the terms page
	pdf.SetFillColor(213, 197, 138) // Same color as the header
	pdf.Rect(0, 0, pageWidth, 20, "F")

	// Add title with black text on the orange background
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0) // Black text
	pdf.SetXY(leftMargin, 5)
	pdf.Cell(contentWidth, 10, "TERMA DAN SYARAT PERKHIDMATAN")

	// Reset position after header
	pdf.SetY(29)
	pdf.SetTextColor(0, 0, 0) // Black text

	// Add introductory paragraph with a subtle background
	pdf.SetFillColor(245, 245, 245) // Light gray background
	pdf.RoundedRect(leftMargin, 26, contentWidth, 20, 3, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetXY(leftMargin+5, 29) // Add padding inside the box
	pdf.MultiCell(contentWidth-10, 4, "Berikut adalah terma dan syarat penggunaan laman web "+ticketGroupName+" bagi pembelian secara dalam talian. Sekiranya anda mengakses laman web ini dan menggunakan perkhidmatan yang ditawarkan, anda bersetuju dan penggunaan anda terhadap bahawa anda terikat kepada terma dan syarat sebagaimana berikut :", "0", "L", false)

	// Position after the intro box
	pdf.SetY(50) // Adjusted position to make content more compact

	// Create section boxes for each section with alternating colors
	sectionY := 50.0 // Start position for sections

	// Section i - with a styled section header
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	pdf.SetFont("Arial", "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)     // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, "i) Pembelian Secara Dalam Talian")

	// Section i content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight := 20.0           // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)   // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, "Pembeli hendaklah memastikan tarikh, hari, jenis tiket dan kuantiti adalah betul sebelum mengikut butang bayaran.\n\nBagi bayaran melalui kad kredit, kad debit atau perkhidmatan perbankan internet seperti Maybank2U atau lain-lain bank, anda hendaklah memastikan anda adalah pemilik akaun dan maklum mengenai pembayaran tersebut.", "0", "L", false)

	// Section ii
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	pdf.SetFont("Arial", "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)     // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, "ii) Pengesahan Pembelian")

	// Section ii content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 18.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)   // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, "Selepas penerimaan pembayaran, anda akan menerima resit dan tiket yang tertera QR Code melalui emel yang telah didaftarkan. Sila bawa bersama resit dan tiket tersebut semasa berkunjung ke "+ticketGroupName+" bagi mengelakkan sebarang permasalahan.", "0", "L", false)

	// Section iii
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	pdf.SetFont("Arial", "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)     // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, "iii) Polisi Bayaran Balik")

	// Section iii content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 38.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)   // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi tiada bayaran balik. Kesemua bayaran yang telah diterima tidak akan dibayar balik kepada pembeli kecuali di dalam keadaan tertentu yang akan dibenarkan oleh pihak pengurusan antaranya permasalahan yang tidak dapat dielakkan seperti masalah teknikal laman web/sistem atau permasalahan berkaitan sistem perbankan.\n\nProses bayaran balik adalah dalam tempoh 14 hari dari tarikh masalah dikenapasti. Bagi situasi di mana pembeli telah terlebih membuat bayaran (sekiarnya ada), bayaran balik hanya akan dilaksanakan setelah pihak pengurusan berhukungan akan kepada pihak pengurusan.", "0", "L", false)

	// Section iv
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	pdf.SetFont("Arial", "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)     // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, "iv) Polisi Menukar Tarikh Tiket")

	// Section iv content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 18.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)   // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi penukaran tarikh adalah tidak dibenarkan. Sekiranya pengunjung tidak dapat hadir pada tarikh yang telah diguatakan, penukaran tarikh tiket adalah tidak dibenarkan dan tiada pulangan bayaran akan dibuat.", "0", "L", false)

	// Section v
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	pdf.SetFont("Arial", "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)     // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, "v) Had Tanggungjawab")

	// Section v content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 30.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	pdf.SetFont("Arial", "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)   // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, "Pihak pengurusan tidak menjamin bahawa fungsi yang terdapat di dalam laman web ini tidak akan terganggu atau bebas dari sebarang kesalahan. Pihak pengurusan juga tidak akan bertanggungjawab atas sebarang kerugian, kemusnahan, gangguran perkhidmatan, kerugian, kehilangan simpanan atau kesan sampingan yang lain ketika mengoperasikan atau kegagalan mengoperasikan laman web ini, akses tanpa kebenaran, kenyataan atau tindakan pihak ketiga di laman web ini atau perkara-perkara lain yang berkaitan dengan laman web.", "0", "L", false)

	// Add Contact Us section at bottom
	sectionY += sectionHeight + 15

	// Contact section header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 6, "Hubungi Kami")
	sectionY += 8

	// Contact details
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 5, "TEL    : "+contactNo)
	sectionY += 6

	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 5, "Emel : "+email)
}

// addOrderDetailsPage adds a page with the order details
func addOrderDetailsPage(pdf *gofpdf.Fpdf, orderOverview email.OrderOverview, orderItems []email.OrderInfo) {
	pdf.AddPage()

	// Set margins for the order details page
	leftMargin := 15.0
	pdf.SetMargins(leftMargin, 20, 15)

	// Add Order Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(51, 51, 51) // Dark gray
	pdf.Cell(170, 20, fmt.Sprintf("Pesanan #%s", orderOverview.OrderNumber))
	pdf.Ln(30)

	// Order Details Section
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(51, 51, 51)

	// Create left column
	col1Width := 40.0
	col2Width := 80.0

	// Placed At
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(col1Width, 10, "Dibuat Pada")

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, orderOverview.PurchaseDate)
	pdf.Ln(10)

	// Total
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(col1Width, 10, "Jumlah")

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, fmt.Sprintf("MYR %.2f", orderOverview.Total))
	pdf.Ln(10)

	// Status
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(col1Width, 10, "Status")

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, "DISAHKAN")
	pdf.Ln(20)

	// Items Table
	// Table headers
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFillColor(213, 197, 138)

	itemColWidth := 70.0
	priceColWidth := 35.0
	qtyColWidth := 40.0
	totalColWidth := 40.0

	// Draw table header with fill color
	pdf.SetFillColor(213, 197, 138)
	pdf.Rect(leftMargin, pdf.GetY(), itemColWidth+priceColWidth+qtyColWidth+totalColWidth, 10, "F")

	// Item header
	pdf.Cell(itemColWidth, 10, "Barangan")

	// Price header
	pdf.Cell(priceColWidth, 10, "Harga")

	// Quantity header
	pdf.Cell(qtyColWidth, 10, "Kuantiti")

	// Total header
	pdf.Cell(totalColWidth, 10, "Jumlah")
	pdf.Ln(10)

	// Table rows
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	// Loop through items to create table rows
	for i, item := range orderItems {
		// If not enough space for this row and the totals section, add a new page
		if pdf.GetY() > 250 {
			pdf.AddPage()
			pdf.SetMargins(leftMargin, 20, 20)
			pdf.SetY(30)
		}

		startY := pdf.GetY()

		// Item name and description
		nameY := startY
		// Add 3mm padding to the left of text in the item column
		pdf.MultiCell(itemColWidth, 6, fmt.Sprintf("%s\n%s", item.Description, item.EntryDate), "", "", false)
		nameHeight := pdf.GetY() - nameY

		// Reset Y position to start of the row for other columns
		pdf.SetY(startY)
		pdf.SetX(leftMargin + itemColWidth)

		// Price
		pdf.Cell(priceColWidth, nameHeight, fmt.Sprintf("MYR %.2f", item.Price))

		// Quantity
		pdf.Cell(qtyColWidth, nameHeight, fmt.Sprintf("%d", item.Quantity))

		// Total for this item
		itemTotal := item.Price * float64(item.Quantity)
		pdf.Cell(totalColWidth, nameHeight, fmt.Sprintf("MYR %.2f", itemTotal))

		// Move to the next line after the tallest cell
		pdf.SetY(startY + nameHeight)

		// Add a line after each row except the last one
		if i < len(orderItems)-1 {
			pdf.SetDrawColor(220, 220, 220) // Light gray
			pdf.Line(leftMargin, pdf.GetY(), leftMargin+itemColWidth+priceColWidth+qtyColWidth+totalColWidth, pdf.GetY())
			pdf.Ln(5) // Add some space after the line
		}
	}

	// Summary Section
	pdf.Ln(15)

	// Line above subtotal
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(leftMargin, pdf.GetY()-5, leftMargin+itemColWidth+priceColWidth+qtyColWidth+totalColWidth, pdf.GetY()-5)

	// Subtotal
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(51, 51, 51)

	// Calculate and position the summary on the right
	summaryX := leftMargin + itemColWidth + priceColWidth
	//summaryWidth := qtyColWidth + totalColWidth

	// Subtotal row
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, "Jumlah Kecil")
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("MYR %.2f", orderOverview.Total))
	pdf.Ln(10)

	// Discount row (always show it as requested)
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, "Jumlah Diskaun")
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("MYR %.2f", 0.00))
	pdf.Ln(10)

	// Total row
	pdf.SetFont("Arial", "B", 11)
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, "Jumlah")

	finalTotal := orderOverview.Total - 0
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("MYR %.2f", finalTotal))

	// Add GST note if applicable
	//if order.GST {
	//	pdf.Ln(6)
	//	pdf.SetFont("Arial", "I", 9)
	//	pdf.SetTextColor(100, 100, 100)
	//	pdf.SetX(summaryX)
	//	pdf.Cell(summaryWidth, 10, "Inclusive GST")
	//}
}

// GenerateTicketPDF generates a PDF ticket with QR codes arranged horizontally
func (s *PDFService) GenerateTicketPDF(orderOverview email.OrderOverview, orderItems []email.OrderInfo, tickets []email.TicketInfo) ([]byte, string, error) {
	var addr1 string
	var addr2 string
	var contactNo string
	var email string
	var logoBase64 string
	if orderOverview.TicketGroup == "Zoo Johor" {
		addr1 = "Jalan Gertak Merah, Taman Istana"
		addr2 = "80000 Johor Bahru, Johor"
		contactNo = "+607-223 0404"
		email = "zoojohor@johor.gov.my"
		logoBase64 = storage.ZooLogo
	} else {
		addr1 = "Taman Botani Diraja Johor Istana Besar Johor"
		addr2 = "80000 Johor Bahru, Johor"
		contactNo = "+607-485 8101"
		email = "botani.johor@gmail.com"
		logoBase64 = storage.BotaniLogo
	}

	// Create a new PDF with portrait orientation, mm unit, A4 format
	pdf := gofpdf.New("P", "mm", "A4", "")

	// QR code settings
	qrSize := 25.0        // Size of QR code in mm
	startX := 15.0        // Starting X position in mm (left margin)
	horizontalGap := 35.0 // Gap between QR codes in mm
	startY := 150.0       // Starting Y position in mm (adjusted to leave space for header, participant info and redeem section)
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

		// Add the header
		addHeader(logoBase64, orderOverview.TicketGroup, addr1, addr2, contactNo, pdf)

		// Add participant information section
		addParticipantInfo(pdf, orderOverview.FullName, orderOverview.PurchaseDate, orderOverview.EntryDate, orderOverview.OrderNumber, orderOverview.Quantity)

		// Add redemption instructions section
		addRedeemSection(pdf)

		// Set default font for ticket content
		pdf.SetFont("Arial", "B", 16)

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
			posY := startY + float64(row)*45.0 // 45mm vertical spacing between rows

			// Generate QR code image
			qrCode, err := qr.Encode(ticket.Content, qr.M, qr.Auto)
			if err != nil {
				fmt.Printf("failed to generate QR code: %w", err)
			}

			// Scale the QR code
			qrCode, err = barcode.Scale(qrCode, 256, 256)
			if err != nil {
				fmt.Printf("failed to scale QR code: %w", err)
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
				fmt.Printf("failed to encode QR code as PNG: %w", err)
			}

			// Register the in-memory image with gofpdf
			// The image ID must be unique for each image
			imageID := fmt.Sprintf("qr_%d", ticketIndex)
			pdf.RegisterImageOptionsReader(imageID, qrOptions, bytes.NewReader(qrBuffer.Bytes()))

			// Add QR code to PDF at calculated position
			pdf.Image(imageID, posX, posY, qrSize, 0, false, "", 0, "")

			// Add label under QR code
			pdf.SetFont("Arial", "", 8) // Smaller font for better fitting
			labelWidth := qrSize + 10   // Make label width slightly wider than QR code
			pdf.SetY(posY + qrSize + labelOffset)
			pdf.SetX(posX - 5)                                         // Center the label by adjusting starting position
			pdf.MultiCell(labelWidth, 4, ticket.Label, "", "C", false) // 4mm line height, centered text

			// Reset to bold font for the next ticket
			pdf.SetFont("Arial", "B", 16)
		}
	}

	// Add Terms and Conditions page
	addTermsAndConditionsPage(orderOverview.TicketGroup, contactNo, email, pdf)

	// Add Order Details page (single order with multiple items)
	addOrderDetailsPage(pdf, orderOverview, orderItems)

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
