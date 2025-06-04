package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/storage"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
)

// PDFService handles PDF generation functionality
type PDFService struct {
	translations map[string]map[string]string
	fontPath     string // Path to font files
}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{
		translations: initTranslations(),
		fontPath:     "pkg/public/fonts/", // Set your font directory path
	}
}

// setupFonts initializes custom fonts for the PDF
func (s *PDFService) setupFonts(pdf *gofpdf.Fpdf) error {
	// For Chinese support, you need to download and include Chinese fonts
	// Download from: https://github.com/googlefonts/noto-cjk/releases
	// Or use other Chinese fonts like SimSun, SimHei, etc.

	// Add UTF-8 fonts for Chinese support
	// Make sure these font files exist in your fonts directory
	cnFontDir := filepath.Join(s.fontPath, "chinese")

	// Add Chinese fonts
	pdf.AddUTF8Font("NotoSansCJK", "", filepath.Join(cnFontDir, "NotoSansSC-Regular.ttf"))
	pdf.AddUTF8Font("NotoSansCJK", "B", filepath.Join(cnFontDir, "NotoSansSC-Bold.ttf"))

	return nil
}

// setFontForLanguage sets the appropriate font based on language
func setFontForLanguage(pdf *gofpdf.Fpdf, lang string, style string, size float64) {
	switch lang {
	case "cn":
		// Use Chinese font
		if style == "B" {
			pdf.SetFont("NotoSansCJK", "B", size)
		} else {
			pdf.SetFont("NotoSansCJK", "", size)
		}
	default:
		// Use default Arial for English and Malay
		pdf.SetFont("Arial", style, size)
	}
}

// initTranslations initializes all translations
func initTranslations() map[string]map[string]string {
	return map[string]map[string]string{
		"bm": {
			// Headers and labels
			"lead_participant": "Peserta Utama",
			"purchase_date":    "Tarikh Pembelian",
			"entry_date":       "Tarikh Masuk",
			"total_tickets":    "Jumlah Tiket",
			"order_no":         "No. Pesanan",
			"general_line":     "Talian Umum",

			// Redeem section
			"redeem_title":       "Tebus Unit Individu",
			"redeem_instruction": "Imbas kod QR di bawah untuk menebus unit anda secara individu.",

			// Terms and conditions
			"terms_title": "TERMA DAN SYARAT PERKHIDMATAN",
			"terms_intro": "Berikut adalah terma dan syarat penggunaan laman web %s bagi pembelian secara dalam talian. Sekiranya anda mengakses laman web ini dan menggunakan perkhidmatan yang ditawarkan, anda bersetuju dan penggunaan anda terhadap bahawa anda terikat kepada terma dan syarat sebagaimana berikut :",

			"section1_title":   "i) Pembelian Secara Dalam Talian",
			"section1_content": "Pembeli hendaklah memastikan tarikh, hari, jenis tiket dan kuantiti adalah betul sebelum mengikut butang bayaran.\n\nBagi bayaran melalui kad kredit, kad debit atau perkhidmatan perbankan internet seperti Maybank2U atau lain-lain bank, anda hendaklah memastikan anda adalah pemilik akaun dan maklum mengenai pembayaran tersebut.",

			"section2_title":   "ii) Pengesahan Pembelian",
			"section2_content": "Selepas penerimaan pembayaran, anda akan menerima resit dan tiket yang tertera QR Code melalui emel yang telah didaftarkan. Sila bawa bersama resit dan tiket tersebut semasa berkunjung ke %s bagi mengelakkan sebarang permasalahan.",

			"section3_title":   "iii) Polisi Bayaran Balik",
			"section3_content": "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi tiada bayaran balik. Kesemua bayaran yang telah diterima tidak akan dibayar balik kepada pembeli kecuali di dalam keadaan tertentu yang akan dibenarkan oleh pihak pengurusan antaranya permasalahan yang tidak dapat dielakkan seperti masalah teknikal laman web/sistem atau permasalahan berkaitan sistem perbankan.\n\nProses bayaran balik adalah dalam tempoh 14 hari dari tarikh masalah dikenapasti. Bagi situasi di mana pembeli telah terlebih membuat bayaran (sekiranya ada), bayaran balik hanya akan dilaksanakan setelah pihak pengurusan berhukungan akan kepada pihak pengurusan.",

			"section4_title":   "iv) Polisi Menukar Tarikh Tiket",
			"section4_content": "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi penukaran tarikh adalah tidak dibenarkan. Sekiranya pengunjung tidak dapat hadir pada tarikh yang telah dijadualkan, penukaran tarikh tiket adalah tidak dibenarkan dan tiada pulangan bayaran akan dibuat.",

			"section5_title":   "v) Had Tanggungjawab",
			"section5_content": "Pihak pengurusan tidak menjamin bahawa fungsi yang terdapat di dalam laman web ini tidak akan terganggu atau bebas dari sebarang kesalahan. Pihak pengurusan juga tidak akan bertanggungjawab atas sebarang kerugian, kemusnahan, gangguan perkhidmatan, kerugian, kehilangan simpanan atau kesan sampingan yang lain ketika mengoperasikan atau kegagalan mengoperasikan laman web ini, akses tanpa kebenaran, kenyataan atau tindakan pihak ketiga di laman web ini atau perkara-perkara lain yang berkaitan dengan laman web.",

			"contact_us": "Hubungi Kami",
			"tel":        "TEL",
			"email":      "Emel",

			// Order details
			"order_title":     "Pesanan",
			"placed_at":       "Dibuat Pada",
			"total":           "Jumlah",
			"status":          "Status",
			"confirmed":       "DISAHKAN",
			"items":           "Barangan",
			"price":           "Harga",
			"quantity":        "Kuantiti",
			"subtotal":        "Jumlah Kecil",
			"discount_amount": "Jumlah Diskaun",
			"currency":        "MYR",
		},
		"en": {
			// Headers and labels
			"lead_participant": "Lead Participant",
			"purchase_date":    "Purchase Date",
			"entry_date":       "Entry Date",
			"total_tickets":    "Total Tickets",
			"order_no":         "Order No.",
			"general_line":     "General Line",

			// Redeem section
			"redeem_title":       "Redeem Individual Units",
			"redeem_instruction": "Scan the QR codes below to redeem your units individually.",

			// Terms and conditions
			"terms_title": "TERMS AND CONDITIONS OF SERVICE",
			"terms_intro": "The following are the terms and conditions for using the %s website for online purchases. If you access this website and use the services offered, you agree and acknowledge that you are bound by the following terms and conditions:",

			"section1_title":   "i) Online Purchase",
			"section1_content": "Buyers must ensure that the date, day, ticket type and quantity are correct before clicking the payment button.\n\nFor payments via credit card, debit card or internet banking services such as Maybank2U or other banks, you must ensure that you are the account owner and are aware of the payment.",

			"section2_title":   "ii) Purchase Confirmation",
			"section2_content": "After payment is received, you will receive a receipt and ticket with QR Code via the registered email. Please bring the receipt and ticket when visiting %s to avoid any problems.",

			"section3_title":   "iii) Refund Policy",
			"section3_content": "This online ticket purchasing service operates on a no-refund policy. All payments received will not be refunded to buyers except in certain circumstances permitted by management, including unavoidable issues such as technical problems with the website/system or banking system issues.\n\nThe refund process is within 14 days from the date the issue is identified. In situations where buyers have made excess payments (if any), refunds will only be processed after management review.",

			"section4_title":   "iv) Ticket Date Change Policy",
			"section4_content": "This online ticket purchasing service operates on a policy where date changes are not allowed. If visitors cannot attend on the scheduled date, ticket date changes are not permitted and no refunds will be made.",

			"section5_title":   "v) Limitation of Liability",
			"section5_content": "Management does not guarantee that the functions on this website will be uninterrupted or error-free. Management will also not be responsible for any damage, destruction, service interruption, loss, loss of savings or other side effects when operating or failing to operate this website, unauthorized access, statements or actions of third parties on this website or other matters related to this website.",

			"contact_us": "Contact Us",
			"tel":        "TEL",
			"email":      "Email",

			// Order details
			"order_title":     "Order",
			"placed_at":       "Placed At",
			"total":           "Total",
			"status":          "Status",
			"confirmed":       "CONFIRMED",
			"items":           "Items",
			"price":           "Price",
			"quantity":        "Quantity",
			"subtotal":        "Subtotal",
			"discount_amount": "Discount Amount",
			"currency":        "MYR",
		},
		"cn": {
			// Headers and labels
			"lead_participant": "主要参与者",
			"purchase_date":    "购买日期",
			"entry_date":       "入场日期",
			"total_tickets":    "门票总数",
			"order_no":         "订单号",
			"general_line":     "总机",

			// Redeem section
			"redeem_title":       "兑换个人单位",
			"redeem_instruction": "扫描下方的二维码以单独兑换您的单位。",

			// Terms and conditions
			"terms_title": "服务条款和条件",
			"terms_intro": "以下是使用%s网站进行在线购买的条款和条件。如果您访问本网站并使用所提供的服务，即表示您同意并承认您受以下条款和条件的约束：",

			"section1_title":   "i) 在线购买",
			"section1_content": "买家必须在点击付款按钮之前确保日期、日期、门票类型和数量正确。\n\n对于通过信用卡、借记卡或网上银行服务（如Maybank2U或其他银行）付款，您必须确保您是账户所有者并了解该付款。",

			"section2_title":   "ii) 购买确认",
			"section2_content": "收到付款后，您将通过注册的电子邮件收到带有二维码的收据和门票。请在访问%s时携带收据和门票，以避免任何问题。",

			"section3_title":   "iii) 退款政策",
			"section3_content": "此在线购票服务采用不退款政策。除管理层允许的某些情况外，所有收到的付款将不会退还给买家，包括不可避免的问题，如网站/系统的技术问题或银行系统问题。\n\n退款流程在确定问题之日起14天内。在买家多付款项的情况下（如有），只有在管理层审核后才会处理退款。",

			"section4_title":   "iv) 门票日期更改政策",
			"section4_content": "此在线购票服务采用不允许更改日期的政策。如果访客无法在预定日期出席，则不允许更改门票日期，也不会退款。",

			"section5_title":   "v) 责任限制",
			"section5_content": "管理层不保证本网站的功能不会中断或无错误。管理层也不对在操作或未能操作本网站时造成的任何损害、破坏、服务中断、损失、储蓄损失或其他副作用负责，未经授权的访问、第三方在本网站上的声明或行为或与本网站相关的其他事项。",

			"contact_us": "联系我们",
			"tel":        "电话",
			"email":      "电子邮件",

			// Order details
			"order_title":     "订单",
			"placed_at":       "下单时间",
			"total":           "总计",
			"status":          "状态",
			"confirmed":       "已确认",
			"items":           "项目",
			"price":           "价格",
			"quantity":        "数量",
			"subtotal":        "小计",
			"discount_amount": "折扣金额",
			"currency":        "MYR",
		},
	}
}

// getTranslation gets a translation for a key in the specified language
func (s *PDFService) getTranslation(lang, key string) string {
	if translations, ok := s.translations[lang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}
	// Fallback to English if translation not found
	if translations, ok := s.translations["en"]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}
	return key // Return key if no translation found
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

func addHeader(logoBase64 string, ticketGroupName string, addr1 string, addr2 string, generalLine string, pdf *gofpdf.Fpdf, lang string, s *PDFService) {
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
	setFontForLanguage(pdf, lang, "B", 16)
	pdf.SetTextColor(0, 0, 0) // Black text

	// Title - use MultiCell for auto-wrapping long text
	titleY := boxY + 5 // Start a bit higher to accommodate potentially two lines
	pdf.SetXY(boxX+5, titleY)
	pdf.MultiCell(boxWidth-10, 8, strings.ToUpper(ticketGroupName), "", "C", false)

	// Get current Y position after title, to make sure address lines don't overlap
	currentY := pdf.GetY() + 3 // Add small gap after title

	// Address lines
	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetXY(boxX, currentY)
	pdf.CellFormat(boxWidth, 8, addr1, "", 0, "C", false, 0, "")

	pdf.SetXY(boxX, currentY+8)
	pdf.CellFormat(boxWidth, 8, addr2, "", 0, "C", false, 0, "")

	// General line
	pdf.SetXY(boxX, currentY+16)
	pdf.CellFormat(boxWidth, 8, s.getTranslation(lang, "general_line")+": "+generalLine, "", 0, "C", false, 0, "")

	// Reset text color to black for the rest of the document
	pdf.SetTextColor(0, 0, 0)
}

// addParticipantInfo adds the participant information section
func addParticipantInfo(pdf *gofpdf.Fpdf, participantName string, purchaseDate string, entryDate string, orderNo string, totalTickets int, lang string, s *PDFService) {
	startY := 85.0 // Start position after the header
	leftColX := 15.0
	midColX := 80.0
	rightColX := 150.0

	// Column titles - Left column
	setFontForLanguage(pdf, lang, "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(leftColX, startY)
	pdf.Cell(100, 5, s.getTranslation(lang, "lead_participant"))

	// Column titles - Middle column
	pdf.SetXY(midColX, startY)
	pdf.Cell(100, 5, s.getTranslation(lang, "purchase_date"))

	// Column titles - Right column
	pdf.SetXY(rightColX, startY)
	pdf.Cell(100, 5, s.getTranslation(lang, "entry_date"))

	// Values - Left column (Participant Name)
	setFontForLanguage(pdf, lang, "B", 12)
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
	setFontForLanguage(pdf, lang, "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(leftColX, startY+18)
	pdf.Cell(100, 5, s.getTranslation(lang, "total_tickets"))

	// Order No. - Label
	setFontForLanguage(pdf, lang, "", 9)
	pdf.SetTextColor(100, 100, 100) // Gray for the labels
	pdf.SetXY(midColX, startY+18)
	pdf.Cell(100, 5, s.getTranslation(lang, "order_no"))

	// Order No. - Value
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black for the values
	pdf.SetXY(midColX, startY+25)
	pdf.Cell(100, 5, orderNo)

	// Total Tickets - Value
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black for the values
	pdf.SetXY(leftColX, startY+25)
	pdf.Cell(100, 5, strconv.Itoa(totalTickets))
}

// addRedeemSection adds the redeem instructions section
func addRedeemSection(pdf *gofpdf.Fpdf, lang string, s *PDFService) {
	startY := 130.0 // Start position after the participant info
	leftX := 15.0

	// Add section heading
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(leftX, startY)
	pdf.Cell(180, 5, s.getTranslation(lang, "redeem_title"))

	// Add instruction text
	setFontForLanguage(pdf, lang, "", 10)
	pdf.SetXY(leftX, startY+7)
	pdf.Cell(180, 5, s.getTranslation(lang, "redeem_instruction"))
}

// addTermsAndConditionsPage adds a new page with Terms and Conditions content
func addTermsAndConditionsPageDeprecated(ticketGroupName string, contactNo string, email string, pdf *gofpdf.Fpdf, lang string, s *PDFService) {
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
	setFontForLanguage(pdf, lang, "B", 14)
	pdf.SetTextColor(0, 0, 0) // Black text
	pdf.SetXY(leftMargin, 5)
	pdf.Cell(contentWidth, 10, s.getTranslation(lang, "terms_title"))

	// Reset position after header
	pdf.SetY(29)
	pdf.SetTextColor(0, 0, 0) // Black text

	// Add introductory paragraph with a subtle background
	pdf.SetFillColor(245, 245, 245) // Light gray background
	pdf.RoundedRect(leftMargin, 26, contentWidth, 20, 3, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetXY(leftMargin+5, 29)          // Add padding inside the box
	introText := fmt.Sprintf(s.getTranslation(lang, "terms_intro"), ticketGroupName)
	pdf.MultiCell(contentWidth-10, 4, introText, "0", "L", false)

	// Position after the intro box
	pdf.SetY(50) // Adjusted position to make content more compact

	// Create section boxes for each section with alternating colors
	sectionY := 50.0 // Start position for sections

	// Section i - with a styled section header
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	setFontForLanguage(pdf, lang, "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)              // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, s.getTranslation(lang, "section1_title"))

	// Section i content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight := 20.0           // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)            // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, s.getTranslation(lang, "section1_content"), "0", "L", false)

	// Section ii
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	setFontForLanguage(pdf, lang, "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)              // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, s.getTranslation(lang, "section2_title"))

	// Section ii content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 18.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)            // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	section2Content := fmt.Sprintf(s.getTranslation(lang, "section2_content"), ticketGroupName)
	pdf.MultiCell(contentWidth-10, 4, section2Content, "0", "L", false)

	// Section iii
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	setFontForLanguage(pdf, lang, "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)              // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, s.getTranslation(lang, "section3_title"))

	// Section iii content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 38.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)            // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, s.getTranslation(lang, "section3_content"), "0", "L", false)

	// Section iv
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	setFontForLanguage(pdf, lang, "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)              // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, s.getTranslation(lang, "section4_title"))

	// Section iv content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 18.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)            // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, s.getTranslation(lang, "section4_content"), "0", "L", false)

	// Section v
	sectionY += sectionHeight + 10                                         // Reduced spacing
	pdf.SetFillColor(213, 197, 138)                                        // Khaki accent color
	pdf.RoundedRect(leftMargin, sectionY, contentWidth, 7, 2, "1234", "F") // Reduced height

	setFontForLanguage(pdf, lang, "B", 10) // Reduced font size
	pdf.SetTextColor(0, 0, 0)              // Black text for section title
	pdf.SetXY(leftMargin+5, sectionY+1.5)
	pdf.Cell(contentWidth-10, 4, s.getTranslation(lang, "section5_title"))

	// Section v content
	pdf.SetFillColor(250, 250, 250) // Very light gray for section content
	sectionHeight = 30.0            // Reduced height
	pdf.RoundedRect(leftMargin, sectionY+7, contentWidth, sectionHeight, 2, "1234", "F")

	setFontForLanguage(pdf, lang, "", 9) // Reduced font size
	pdf.SetTextColor(0, 0, 0)            // Black text for content
	pdf.SetXY(leftMargin+5, sectionY+9)
	pdf.MultiCell(contentWidth-10, 4, s.getTranslation(lang, "section5_content"), "0", "L", false)

	// Add Contact Us section at bottom
	sectionY += sectionHeight + 15

	// Contact section header
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 6, s.getTranslation(lang, "contact_us"))
	sectionY += 8

	// Contact details
	setFontForLanguage(pdf, lang, "", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "tel")+"    : "+contactNo)
	sectionY += 6

	pdf.SetXY(leftMargin, sectionY)
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "email")+" : "+email)
}

// addTermsAndConditionsPage adds a new page with improved Terms and Conditions layout
func addTermsAndConditionsPage(ticketGroupName string, contactNo string, email string, pdf *gofpdf.Fpdf, lang string, s *PDFService) {
	pdf.AddPage()

	// Set margins for the terms page
	leftMargin := 15.0
	topMargin := 15.0
	rightMargin := 15.0
	pdf.SetMargins(leftMargin, topMargin, rightMargin)

	// Page width calculation for design elements
	pageWidth := 210.0                                   // A4 width in mm
	contentWidth := pageWidth - leftMargin - rightMargin // Left and right margins

	// Add a stylish header bar for the terms page
	pdf.SetFillColor(213, 197, 138) // Same color as the header
	pdf.Rect(0, 0, pageWidth, 20, "F")

	// Add title with black text on the orange background
	setFontForLanguage(pdf, lang, "B", 14)
	pdf.SetTextColor(0, 0, 0) // Black text
	pdf.SetXY(leftMargin, 5)
	pdf.Cell(contentWidth, 10, s.getTranslation(lang, "terms_title"))

	// Reset position after header
	pdf.SetY(25)
	pdf.SetTextColor(0, 0, 0) // Black text

	// Add introductory paragraph
	setFontForLanguage(pdf, lang, "", 9)
	introText := fmt.Sprintf(s.getTranslation(lang, "terms_intro"), ticketGroupName)
	pdf.MultiCell(contentWidth, 5, introText, "", "L", false)

	// Add spacing after intro
	pdf.Ln(5)

	// Define sections - each section has a title and content
	sections := []struct {
		titleKey   string
		contentKey string
		// For sections that need formatting with ticketGroupName
		needsFormatting bool
	}{
		{
			titleKey:        "section1_title",
			contentKey:      "section1_content",
			needsFormatting: false,
		},
		{
			titleKey:        "section2_title",
			contentKey:      "section2_content",
			needsFormatting: true, // This section needs ticketGroupName
		},
		{
			titleKey:        "section3_title",
			contentKey:      "section3_content",
			needsFormatting: false,
		},
		{
			titleKey:        "section4_title",
			contentKey:      "section4_content",
			needsFormatting: false,
		},
		{
			titleKey:        "section5_title",
			contentKey:      "section5_content",
			needsFormatting: false,
		},
	}

	// Process each section with dynamic spacing
	for _, section := range sections {
		// Check if we need a new page
		// Reserve space for contact section at bottom (approximately 40mm)
		if pdf.GetY() > 240 {
			pdf.AddPage()
			pdf.SetY(topMargin)
		}

		// Get current Y position before section
		startY := pdf.GetY()

		// Add section title
		setFontForLanguage(pdf, lang, "B", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.MultiCell(contentWidth, 6, s.getTranslation(lang, section.titleKey), "", "L", false)

		// Add a small gap between title and content
		pdf.Ln(2)

		// Add section content
		setFontForLanguage(pdf, lang, "", 9)
		pdf.SetTextColor(0, 0, 0)

		// Get content text
		contentText := s.getTranslation(lang, section.contentKey)
		if section.needsFormatting {
			contentText = fmt.Sprintf(contentText, ticketGroupName)
		}

		// Write content
		pdf.MultiCell(contentWidth, 5, contentText, "", "L", false)

		// Calculate the height used by this section
		endY := pdf.GetY()
		sectionHeight := endY - startY

		// Add dynamic spacing based on section height
		// Smaller sections get more spacing, larger sections get less
		var spacing float64
		if sectionHeight < 20 {
			spacing = 8
		} else if sectionHeight < 30 {
			spacing = 6
		} else {
			spacing = 4
		}

		pdf.Ln(spacing)
	}

	// Add Contact Us section at bottom
	// Check if we have enough space for contact section (about 20mm)
	if pdf.GetY() > 260 {
		pdf.AddPage()
		pdf.SetY(topMargin)
	}

	// Add some spacing before contact section
	pdf.Ln(10)

	// Contact section header
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(contentWidth, 6, s.getTranslation(lang, "contact_us"))
	pdf.Ln(8)

	// Contact details
	setFontForLanguage(pdf, lang, "", 10)
	pdf.SetTextColor(51, 51, 51)

	// TEL
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "tel")+"    : "+contactNo)
	pdf.Ln(6)

	// Email
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "email")+" : "+email)
}

// Alternative version with even more dynamic spacing based on actual content measurement
func addTermsAndConditionsPageAdvanced(ticketGroupName string, contactNo string, email string, pdf *gofpdf.Fpdf, lang string, s *PDFService) {
	pdf.AddPage()

	// Set margins for the terms page
	leftMargin := 15.0
	topMargin := 15.0
	rightMargin := 15.0
	pdf.SetMargins(leftMargin, topMargin, rightMargin)

	// Page width calculation
	pageWidth := 210.0
	contentWidth := pageWidth - leftMargin - rightMargin

	// Add header bar
	pdf.SetFillColor(213, 197, 138)
	pdf.Rect(0, 0, pageWidth, 20, "F")

	// Add title
	setFontForLanguage(pdf, lang, "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(leftMargin, 5)
	pdf.Cell(contentWidth, 10, s.getTranslation(lang, "terms_title"))

	// Reset position
	pdf.SetY(25)
	pdf.SetTextColor(0, 0, 0)

	// Add intro paragraph
	setFontForLanguage(pdf, lang, "", 9)
	introText := fmt.Sprintf(s.getTranslation(lang, "terms_intro"), ticketGroupName)
	pdf.MultiCell(contentWidth, 5, introText, "", "L", false)
	pdf.Ln(5)

	// Get total available height for content (excluding contact section)
	pageHeight := 297.0                                                    // A4 height
	contactSectionHeight := 30.0                                           // Estimated height for contact section
	availableHeight := pageHeight - pdf.GetY() - contactSectionHeight - 20 // 20mm bottom margin

	// Calculate content for all sections first to determine spacing
	type sectionInfo struct {
		titleKey        string
		contentKey      string
		needsFormatting bool
		estimatedHeight float64
	}

	sections := []sectionInfo{
		{titleKey: "section1_title", contentKey: "section1_content", needsFormatting: false},
		{titleKey: "section2_title", contentKey: "section2_content", needsFormatting: true},
		{titleKey: "section3_title", contentKey: "section3_content", needsFormatting: false},
		{titleKey: "section4_title", contentKey: "section4_content", needsFormatting: false},
		{titleKey: "section5_title", contentKey: "section5_content", needsFormatting: false},
	}

	// Estimate heights for each section
	totalContentHeight := 0.0
	for i := range sections {
		// Title height (approximately 2 lines with spacing)
		titleHeight := 8.0

		// Content height estimation
		contentText := s.getTranslation(lang, sections[i].contentKey)
		if sections[i].needsFormatting {
			contentText = fmt.Sprintf(contentText, ticketGroupName)
		}

		// Rough estimation: ~80 characters per line, 5mm per line
		contentLines := float64(len(contentText)) / 80.0
		contentHeight := contentLines * 5.0

		sections[i].estimatedHeight = titleHeight + contentHeight + 2 // 2mm gap between title and content
		totalContentHeight += sections[i].estimatedHeight
	}

	// Calculate dynamic spacing between sections
	totalSpacing := availableHeight - totalContentHeight
	if totalSpacing < 0 {
		totalSpacing = 20 // Minimum total spacing if content overflows
	}
	spacingPerSection := totalSpacing / float64(len(sections)-1)
	if spacingPerSection > 10 {
		spacingPerSection = 10 // Cap maximum spacing
	}
	if spacingPerSection < 3 {
		spacingPerSection = 3 // Minimum spacing
	}

	// Now render sections with calculated spacing
	for i, section := range sections {
		// Check for page break
		if pdf.GetY() > 240 {
			pdf.AddPage()
			pdf.SetY(topMargin)
		}

		// Section title
		setFontForLanguage(pdf, lang, "B", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.MultiCell(contentWidth, 6, s.getTranslation(lang, section.titleKey), "", "L", false)
		pdf.Ln(2)

		// Section content
		setFontForLanguage(pdf, lang, "", 9)
		contentText := s.getTranslation(lang, section.contentKey)
		if section.needsFormatting {
			contentText = fmt.Sprintf(contentText, ticketGroupName)
		}
		pdf.MultiCell(contentWidth, 5, contentText, "", "L", false)

		// Add calculated spacing (except after last section)
		if i < len(sections)-1 {
			pdf.Ln(spacingPerSection)
		}
	}

	// Contact section
	if pdf.GetY() > 260 {
		pdf.AddPage()
		pdf.SetY(topMargin)
	}

	pdf.Ln(10)

	// Contact header
	setFontForLanguage(pdf, lang, "B", 12)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(contentWidth, 6, s.getTranslation(lang, "contact_us"))
	pdf.Ln(8)

	// Contact details
	setFontForLanguage(pdf, lang, "", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "tel")+"    : "+contactNo)
	pdf.Ln(6)
	pdf.Cell(contentWidth, 5, s.getTranslation(lang, "email")+" : "+email)
}

// addOrderDetailsPage adds a page with the order details
func addOrderDetailsPage(pdf *gofpdf.Fpdf, orderOverview email.OrderOverview, orderItems []email.OrderInfo, lang string, s *PDFService) {
	pdf.AddPage()

	// Set margins for the order details page
	leftMargin := 15.0
	pdf.SetMargins(leftMargin, 20, 15)

	// Add Order Title
	setFontForLanguage(pdf, lang, "B", 24)
	pdf.SetTextColor(51, 51, 51) // Dark gray
	pdf.Cell(170, 20, fmt.Sprintf("%s #%s", s.getTranslation(lang, "order_title"), orderOverview.OrderNumber))
	pdf.Ln(30)

	// Order Details Section
	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(51, 51, 51)

	// Create left column
	col1Width := 40.0
	col2Width := 80.0

	// Placed At
	setFontForLanguage(pdf, lang, "", 11)
	pdf.Cell(col1Width, 10, s.getTranslation(lang, "placed_at"))

	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, orderOverview.PurchaseDate)
	pdf.Ln(10)

	// Total
	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(col1Width, 10, s.getTranslation(lang, "total"))

	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), orderOverview.Total))
	pdf.Ln(10)

	// Status
	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(col1Width, 10, s.getTranslation(lang, "status"))

	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(col2Width, 10, s.getTranslation(lang, "confirmed"))
	pdf.Ln(20)

	// Items Table
	// Table headers
	setFontForLanguage(pdf, lang, "", 11)
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
	pdf.Cell(itemColWidth, 10, s.getTranslation(lang, "items"))

	// Price header
	pdf.Cell(priceColWidth, 10, s.getTranslation(lang, "price"))

	// Quantity header
	pdf.Cell(qtyColWidth, 10, s.getTranslation(lang, "quantity"))

	// Total header
	pdf.Cell(totalColWidth, 10, s.getTranslation(lang, "total"))
	pdf.Ln(10)

	// Table rows
	setFontForLanguage(pdf, lang, "", 11)
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
		pdf.Cell(priceColWidth, nameHeight, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), item.Price))

		// Quantity
		pdf.Cell(qtyColWidth, nameHeight, fmt.Sprintf("%d", item.Quantity))

		// Total for this item
		itemTotal := item.Price * float64(item.Quantity)
		pdf.Cell(totalColWidth, nameHeight, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), itemTotal))

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
	setFontForLanguage(pdf, lang, "", 11)
	pdf.SetTextColor(51, 51, 51)

	// Calculate and position the summary on the right
	summaryX := leftMargin + itemColWidth + priceColWidth
	//summaryWidth := qtyColWidth + totalColWidth

	// Subtotal row
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, s.getTranslation(lang, "subtotal"))
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), orderOverview.Total))
	pdf.Ln(10)

	// Discount row (always show it as requested)
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, s.getTranslation(lang, "discount_amount"))
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), 0.00))
	pdf.Ln(10)

	// Total row
	setFontForLanguage(pdf, lang, "B", 11)
	pdf.SetX(summaryX)
	pdf.Cell(qtyColWidth, 10, s.getTranslation(lang, "total"))

	finalTotal := orderOverview.Total - 0
	pdf.Cell(totalColWidth, 10, fmt.Sprintf("%s %.2f", s.getTranslation(lang, "currency"), finalTotal))
}

// GenerateTicketPDF generates a PDF ticket with QR codes arranged horizontally
// lang parameter accepts: "bm" (Malay), "en" (English), "cn" (Chinese)
func (s *PDFService) GenerateTicketPDF(orderOverview email.OrderOverview, orderItems []email.OrderInfo, tickets []email.TicketInfo, lang string) ([]byte, string, error) {
	// Validate language parameter
	if lang != "bm" && lang != "en" && lang != "cn" {
		lang = "bm" // Default to Malay if invalid language
	}

	var addr1 string
	var addr2 string
	var contactNo string
	var emailAddr string
	var logoBase64 string

	if orderOverview.TicketGroup == "Zoo Johor" {
		addr1 = "Jalan Gertak Merah, Taman Istana"
		addr2 = "80000 Johor Bahru, Johor"
		contactNo = "+607-223 0404"
		emailAddr = "zoojohor@johor.gov.my"
		logoBase64 = storage.ZooLogo
	} else {
		addr1 = "Taman Botani Diraja Johor Istana Besar Johor"
		addr2 = "80000 Johor Bahru, Johor"
		contactNo = "+607-485 8101"
		emailAddr = "botani.johor@gmail.com"
		logoBase64 = storage.BotaniLogo
	}

	// Create a new PDF with portrait orientation, mm unit, A4 format
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Setup fonts for Chinese support if needed
	if lang == "cn" {
		// Option 1: Use built-in CJK fonts (simpler but limited)
		//pdf.AddUTF8Font("NotoSansCJK", "", "")
		//pdf.AddUTF8Font("NotoSansCJK", "B", "")

		// Option 2: If you have font files, uncomment and use this:
		err := s.setupFonts(pdf)
		if err != nil {
			return nil, "", fmt.Errorf("failed to setup fonts: %w", err)
		}
	}

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
		addHeader(logoBase64, orderOverview.TicketGroup, addr1, addr2, contactNo, pdf, lang, s)

		// Add participant information section
		addParticipantInfo(pdf, orderOverview.FullName, orderOverview.PurchaseDate, orderOverview.EntryDate, orderOverview.OrderNumber, orderOverview.Quantity, lang, s)

		// Add redemption instructions section
		addRedeemSection(pdf, lang, s)

		// Set default font for ticket content
		setFontForLanguage(pdf, lang, "B", 16)

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
			setFontForLanguage(pdf, lang, "", 8) // Smaller font for better fitting
			labelWidth := qrSize + 10            // Make label width slightly wider than QR code
			pdf.SetY(posY + qrSize + labelOffset)
			pdf.SetX(posX - 5)                                         // Center the label by adjusting starting position
			pdf.MultiCell(labelWidth, 4, ticket.Label, "", "C", false) // 4mm line height, centered text

			// Reset to bold font for the next ticket
			setFontForLanguage(pdf, lang, "B", 16)
		}
	}

	// Add Terms and Conditions page
	addTermsAndConditionsPageAdvanced(orderOverview.TicketGroup, contactNo, emailAddr, pdf, lang, s)

	// Add Order Details page (single order with multiple items)
	addOrderDetailsPage(pdf, orderOverview, orderItems, lang, s)

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
