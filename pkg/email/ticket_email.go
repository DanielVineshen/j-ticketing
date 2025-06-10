package email

import (
	"bytes"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"image/png"
	logger "log/slog"
	"strconv"
	"time"
)

// EmailText holds all text strings for different languages
type EmailText struct {
	Subject                 string
	Greeting                string
	ThankYouMessage         string
	LeadParticipant         string
	PurchaseDate            string
	EntryDate               string
	OrderNumber             string
	RedeemTitle             string
	RedeemDescription       string
	ShowQRMessage           string
	EnjoyVisitMessage       string
	TermsTitle              string
	TermsIntro              string
	OnlinePurchaseTitle     string
	OnlinePurchaseContent   string
	PurchaseConfirmTitle    string
	PurchaseConfirmContent  string
	RefundPolicyTitle       string
	RefundPolicyContent     string
	DateChangePolicyTitle   string
	DateChangePolicyContent string
	LiabilityTitle          string
	LiabilityContent        string
	OrderSummaryTitle       string
	ItemsHeader             string
	QuantityHeader          string
	PriceHeader             string
	TotalHeader             string
	SubtotalLabel           string
	TotalWithGSTLabel       string
	AutomaticMessage        string
	ContactUs               string
	Copyright               string
}

// getEmailTexts returns localized text based on language
func getEmailTexts(lang string, ticketGroup string) EmailText {
	switch lang {
	case "en":
		return EmailText{
			Subject:                 fmt.Sprintf("Your %s Tickets", ticketGroup),
			Greeting:                "Dear %s,",
			ThankYouMessage:         fmt.Sprintf("Thank you for your purchase. Here are your tickets for %s:", ticketGroup),
			LeadParticipant:         "Lead Participant",
			PurchaseDate:            "Purchase Date",
			EntryDate:               "Entry Date",
			OrderNumber:             "Order No.",
			RedeemTitle:             "Redeem Individual Units",
			RedeemDescription:       "Scan the QR codes below to redeem your units individually.",
			ShowQRMessage:           "Please show this QR code at the entrance for scanning.",
			EnjoyVisitMessage:       fmt.Sprintf("We hope you enjoy your visit to %s!", ticketGroup),
			TermsTitle:              "Terms and Conditions of Service",
			TermsIntro:              fmt.Sprintf("The following are the terms and conditions for using the %s website for online purchases. If you access this website and use the services offered, it constitutes acknowledgment and agreement that you are bound by the following terms and conditions:", ticketGroup),
			OnlinePurchaseTitle:     "i) Online Purchase",
			OnlinePurchaseContent:   "Buyers must ensure that the date, day, ticket type and quantity are correct before clicking the payment button.\n\nFor payments via credit card, debit card or internet banking services such as Maybank2U or other banks, you must ensure that you are the account owner and are aware of the payment.",
			PurchaseConfirmTitle:    "ii) Purchase Confirmation",
			PurchaseConfirmContent:  fmt.Sprintf("After payment is received, you will receive a receipt and ticket with QR Code via the registered email. Please bring the receipt and ticket with you when visiting %s to avoid any problems.", ticketGroup),
			RefundPolicyTitle:       "iii) Refund Policy",
			RefundPolicyContent:     "This online ticket purchase service operates on a no-refund policy. All payments received will not be refunded to buyers except under certain circumstances determined by management, including unavoidable problems such as website/system technical issues or banking system-related problems.\n\nThe refund process takes 14 days from the date the problem is identified. For situations where buyers have made overpayments (if any), refunds will only be made after proof of payment is submitted to management.",
			DateChangePolicyTitle:   "iv) Ticket Date Change Policy",
			DateChangePolicyContent: "This online ticket purchase service operates on a policy where date changes are not allowed. If visitors cannot attend on the scheduled date, date changes are not permitted and no refund will be made.",
			LiabilityTitle:          "v) Limitation of Liability",
			LiabilityContent:        "Management does not guarantee that the functions contained in this website will not be interrupted or free from any errors. Management will also not be responsible for any damage, destruction, service interruption, loss, loss of savings or other side effects when operating or failing to operate this website, unauthorized access, statements or actions of third parties on this website or other matters related to this website.",
			OrderSummaryTitle:       "Order Summary",
			ItemsHeader:             "Items",
			QuantityHeader:          "Quantity",
			PriceHeader:             "Price",
			TotalHeader:             "Total",
			SubtotalLabel:           "Subtotal",
			TotalWithGSTLabel:       "Total (Including GST)",
			AutomaticMessage:        "This is an automatic message, please do not reply to this email.",
			ContactUs:               "Contact Us:",
			Copyright:               fmt.Sprintf("© %d %s", time.Now().Year(), ticketGroup),
		}
	case "cn":
		return EmailText{
			//Subject:                 fmt.Sprintf("您的%s门票", ticketGroup),
			Subject:                 fmt.Sprintf("Your %s Tickets", ticketGroup),
			Greeting:                "亲爱的%s，",
			ThankYouMessage:         fmt.Sprintf("感谢您的购买。以下是您的%s门票：", ticketGroup),
			LeadParticipant:         "主要参与者",
			PurchaseDate:            "购买日期",
			EntryDate:               "入场日期",
			OrderNumber:             "订单号",
			RedeemTitle:             "兑换个人单位",
			RedeemDescription:       "扫描下面的二维码以单独兑换您的单位。",
			ShowQRMessage:           "请在入口处出示此二维码进行扫描。",
			EnjoyVisitMessage:       fmt.Sprintf("我们希望您享受在%s的参观！", ticketGroup),
			TermsTitle:              "服务条款和条件",
			TermsIntro:              fmt.Sprintf("以下是使用%s网站进行在线购买的条款和条件。如果您访问此网站并使用所提供的服务，即构成承认和同意您受以下条款和条件的约束：", ticketGroup),
			OnlinePurchaseTitle:     "i) 在线购买",
			OnlinePurchaseContent:   "买家必须确保日期、日期、票种和数量正确后再点击付款按钮。\n\n对于通过信用卡、借记卡或网上银行服务（如Maybank2U或其他银行）付款，您必须确保您是账户持有人并了解付款情况。",
			PurchaseConfirmTitle:    "ii) 购买确认",
			PurchaseConfirmContent:  fmt.Sprintf("收到付款后，您将通过注册邮箱收到带有二维码的收据和门票。请在参观%s时携带收据和门票，以避免任何问题。", ticketGroup),
			RefundPolicyTitle:       "iii) 退款政策",
			RefundPolicyContent:     "此在线购票服务采用不退款政策。所有收到的付款将不会退还给买家，除非在管理层确定的某些情况下，包括不可避免的问题，如网站/系统技术问题或银行系统相关问题。\n\n退款流程需要从发现问题之日起14天。对于买家多付款的情况（如果有），只有在向管理层提交付款证明后才会退款。",
			DateChangePolicyTitle:   "iv) 门票日期更改政策",
			DateChangePolicyContent: "此在线购票服务采用不允许更改日期的政策。如果访客无法在预定日期参加，不允许更改日期，也不会退款。",
			LiabilityTitle:          "v) 责任限制",
			LiabilityContent:        "管理层不保证此网站包含的功能不会被中断或免于任何错误。管理层也不会对在操作或无法操作此网站时的任何损害、破坏、服务中断、损失、储蓄损失或其他副作用，未经授权的访问、第三方在此网站上的声明或行为或与此网站相关的其他事项承担责任。",
			OrderSummaryTitle:       "订单摘要",
			ItemsHeader:             "项目",
			QuantityHeader:          "数量",
			PriceHeader:             "价格",
			TotalHeader:             "总计",
			SubtotalLabel:           "小计",
			TotalWithGSTLabel:       "总计（含消费税）",
			AutomaticMessage:        "这是自动消息，请勿回复此邮件。",
			ContactUs:               "联系我们：",
			Copyright:               fmt.Sprintf("© %d %s", time.Now().Year(), ticketGroup),
		}
	default: // "bm" - Bahasa Malaysia (default)
		return EmailText{
			Subject:                 fmt.Sprintf("Tiket %s Anda", ticketGroup),
			Greeting:                "Kepada %s,",
			ThankYouMessage:         fmt.Sprintf("Terima kasih atas pembelian anda. Berikut adalah tiket anda untuk %s:", ticketGroup),
			LeadParticipant:         "Peserta Utama",
			PurchaseDate:            "Tarikh Pembelian",
			EntryDate:               "Tarikh Masuk",
			OrderNumber:             "No. Pesanan",
			RedeemTitle:             "Tebus Unit Individu",
			RedeemDescription:       "Imbas kod QR di bawah untuk menebus unit anda secara individu.",
			ShowQRMessage:           "Sila tunjukkan kod QR ini di pintu masuk untuk imbasan.",
			EnjoyVisitMessage:       fmt.Sprintf("Kami berharap anda menikmati lawatan anda ke %s!", ticketGroup),
			TermsTitle:              "Terma dan Syarat Perkhidmatan",
			TermsIntro:              fmt.Sprintf("Berikut adalah terma dan syarat penggunaan laman web %s bagi pembelian secara dalam talian. Sekiranya anda mengakses laman web ini dan menggunakan perkhidmatan yang ditawarkan, ia merupakan pengakuan dan persetujuan bahawa anda terikat kepada terma dan syarat sebagaimana berikut:", ticketGroup),
			OnlinePurchaseTitle:     "i) Pembelian Secara Dalam Talian",
			OnlinePurchaseContent:   "Pembeli hendaklah memastikan tarikh, hari, jenis tiket dan kuantiti adalah betul sebelum mengklik butang bayaran.\n\nBagi bayaran melalui kad kredit, kad debit atau perkhidmatan perbankan internet seperti Maybank2U atau lain-lain bank, anda hendaklah memastikan anda adalah pemilik akaun dan maklum mengenai pembayaran tersebut.",
			PurchaseConfirmTitle:    "ii) Pengesahan Pembelian",
			PurchaseConfirmContent:  fmt.Sprintf("Selepas penerimaan pembayaran, anda akan menerima resit dan tiket yang tertera QR Code melalui emel yang telah didaftarkan. Sila bawa bersama resit dan tiket tersebut semasa berkunjung ke %s bagi mengelakkan sebarang permasalahan.", ticketGroup),
			RefundPolicyTitle:       "iii) Polisi Bayaran Balik",
			RefundPolicyContent:     "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi tiada bayaran balik. Kesemua bayaran yang telah diterima tidak akan dibayar balik kepada pembeli kecuali di dalam keadaan tertentu yang akan ditentukan oleh pihak pengurusan antaranya permasalahan yang tidak dapat dielakkan seperti masalah teknikal laman web/sistem atau permasalahan berkaitan sistem perbankan.\n\nProses bayaran balik adalah dalam tempoh 14 hari dari tarikh masalah dikenalpasti. Bagi situasi di mana pembeli telah terlebih membuat bayaran (sekiranya ada), bayaran balik hanya akan dilaksanakan setelah bukti pembayaran dikemukakan kepada pihak pengurusan.",
			DateChangePolicyTitle:   "iv) Polisi Menukar Tarikh Tiket",
			DateChangePolicyContent: "Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi penukaran tarikh adalah tidak dibenarkan. Sekiranya pengunjung tidak dapat hadir pada tarikh yang telah dijadualkan, penukaran tarikh adalah tidak dibenarkan dan tiada pulangan bayaran akan dibuat.",
			LiabilityTitle:          "v) Had Tanggungjawab",
			LiabilityContent:        "Pihak pengurusan tidak menjamin bahawa fungsi yang terdapat di dalam laman web ini tidak akan terganggu atau bebas dari sebarang kesalahan. Pihak pengurusan juga tidak akan bertanggungjawab atas sebarang kerosakan, kemusnahan, gangguan perkhidmatan, kerugian, kehilangan simpanan atau kesan sampingan yang lain ketika mengoperasikan atau kegagalan mengoperasikan laman web ini, akses tanpa kebenaran, kenyataan atau tindakan pihak ketiga di laman web ini atau perkara-perkara lain yang berkaitan dengan laman web ini.",
			OrderSummaryTitle:       "Ringkasan Pesanan",
			ItemsHeader:             "Barangan",
			QuantityHeader:          "Kuantiti",
			PriceHeader:             "Harga",
			TotalHeader:             "Jumlah",
			SubtotalLabel:           "Jumlah Kecil",
			TotalWithGSTLabel:       "Jumlah (Termasuk GST)",
			AutomaticMessage:        "Ini adalah mesej automatik, sila jangan balas e-mel ini.",
			ContactUs:               "Hubungi Kami:",
			Copyright:               fmt.Sprintf("© %d %s", time.Now().Year(), ticketGroup),
		}
	}
}

// SendTicketsEmail sends an email with QR codes for tickets in the specified language
func sendTicketsEmail(orderOverview OrderOverview, orderItems []OrderInfo, tickets []TicketInfo, language string) (subjectReturn string, bodyReturn string, attachments []Attachment, err error) {
	// Get localized text
	text := getEmailTexts(language, orderOverview.TicketGroup)

	// Set subject
	subject := text.Subject

	var address string
	var contactNo string
	var email string
	if orderOverview.TicketGroup == "Zoo Johor" {
		if language == "en" {
			address = "Jalan Gertak Merah, Taman Istana<br>80000 Johor Bahru, Johor<br>General Line: +607-223 0404"
		} else if language == "cn" {
			address = "Jalan Gertak Merah, Taman Istana<br>80000 Johor Bahru, Johor<br>总机：+607-223 0404"
		} else {
			address = "Jalan Gertak Merah, Taman Istana<br>80000 Johor Bahru, Johor<br>Talian Umum: +607-223 0404"
		}
		contactNo = "+607-223 0404"
		email = "zoojohor@johor.gov.my"
	} else {
		if language == "en" {
			address = "Taman Botani Diraja Johor Istana Besar Johor<br>80000 Johor Bahru, Johor<br>General Line: +607-223 3020"
		} else if language == "cn" {
			address = "Taman Botani Diraja Johor Istana Besar Johor<br>80000 Johor Bahru, Johor<br>总机：+607-223 3020"
		} else {
			address = "Taman Botani Diraja Johor Istana Besar Johor<br>80000 Johor Bahru, Johor<br>Talian Umum: +607-223 3020"
		}
		contactNo = "+607-485 8101"
		email = "botani.johor@gmail.com"
	}

	// Generate QR code attachments
	var qrAttachments []Attachment
	// Begin building HTML content
	var contentBuilder bytes.Buffer

	contentBuilder.WriteString(fmt.Sprintf(`
    <div style="padding: 20px 0px">
        <h4 style="font-size:16px">%s</h4>
        <p style="font-size:14px">%s</p>
	</div>
    `, fmt.Sprintf(text.Greeting, orderOverview.FullName), text.ThankYouMessage))

	contentBuilder.WriteString(fmt.Sprintf(`
<div class="order-info-section">
    <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 25px 0 30px;">
        <tr>
            <td width="49%%" valign="top">
                <div class="info-group">
                    <div class="info-label">%s</div>
                    <div class="info-value">%s</div>
                </div>
            </td>
            <td width="49%%" valign="top">
                <div class="info-group">
                    <div class="info-label">%s</div>
                    <div class="info-value">%s</div>
                </div>
            </td>
        </tr>
        <tr>
			<td width="49%%" valign="top" style="padding-top: 20px;">
                <div class="info-group">
                    <div class="info-label">%s</div>
                    <div class="info-value">%s</div>
                </div>
            </td>
            <td width="49%%" valign="top" style="padding-top: 20px;">
                <div class="info-group">
                    <div class="info-label">%s</div>
                    <div class="info-value">%s</div>
                </div>
            </td>
        </tr>
    </table>
</div>
`, text.LeadParticipant, orderOverview.FullName, text.PurchaseDate, orderOverview.PurchaseDate,
		text.EntryDate, orderOverview.EntryDate, text.OrderNumber, orderOverview.OrderNumber))

	contentBuilder.WriteString(fmt.Sprintf(`
    <div>
        <div class="redeem-title">
            <h4>%s</h4>
            <p>%s</p>
        </div>
	</div>
    `, text.RedeemTitle, text.RedeemDescription))

	// QR codes section (same as before)
	contentBuilder.WriteString(`
        <table cellspacing="10" cellpadding="0" border="0" align="center" style="margin: 20px auto;">
        <tr>
    `)

	ticketCount := len(tickets)
	maxColumns := 3

	for i, ticket := range tickets {
		// Generate QR code as bytes
		qrBytes, err := generateQRCodeBytes(ticket.Content)
		if err != nil {
			logger.Error("Failed to generate QR code", "ticket", i, "error", err)
			// Add placeholder for failed QR codes
			contentBuilder.WriteString(`
            <td align="center" valign="top" style="padding: 5px; width: 160px;">
                <div style="width: 150px; height: 150px; border: 2px solid #ff6b6b; background: #fff5f5; display: flex; align-items: center; justify-content: center; font-size: 12px; color: #e74c3c;">
                    QR Error
                </div>
                <div style="font-size: 12px; font-weight: bold; color: #333; text-align: center; word-wrap: break-word; line-height: 1.2;">` + ticket.Label + `</div>
            </td>
            `)
			continue
		}

		// Create attachment with Content-ID
		cidName := fmt.Sprintf("qr-code-%d", i)
		qrAttachment := Attachment{
			Name:    fmt.Sprintf("qr_%s.png", ticket.Label),
			Content: qrBytes,
			Type:    "image/png",
			CID:     cidName,
		}
		qrAttachments = append(qrAttachments, qrAttachment)

		// Reference the attachment using cid: in HTML
		contentBuilder.WriteString(`
            <td align="center" valign="top" style="padding: 5px; width: 160px;">
                <img src="cid:` + cidName + `" alt="QR Code" style="width: 150px; height: 150px; border: 1px solid #eee; padding: 5px; margin-bottom: 8px;">
                <div style="font-size: 12px; font-weight: bold; color: #333; text-align: center; word-wrap: break-word; line-height: 1.2;">` + ticket.Label + `</div>
            </td>
        `)

		// Handle row wrapping
		if (i+1)%maxColumns == 0 && i < ticketCount-1 {
			contentBuilder.WriteString(`
			</tr>
			<tr>
			`)
		}
	}

	// Close the QR codes table
	contentBuilder.WriteString(`
        </tr>
        </table>
    `)

	contentBuilder.WriteString(`
        </tr>
        </table>
    `)

	contentBuilder.WriteString(fmt.Sprintf(`
		<div style="text-align: center;">
			<p>%s</p>
        	<p>%s</p>
		</div>
        
        <div class="terms-section">
            <h4>%s</h4>
            <p>%s</p>
            
            <p><strong>%s</strong><br>
            %s</p>
            
            <p><strong>%s</strong><br>
            %s</p>
            
            <p><strong>%s</strong><br>
            %s</p>

			<p><strong>%s</strong><br>
            %s</p>

			<p><strong>%s</strong><br>
            %s</p>
        </div>
    </div>
    `, text.ShowQRMessage, text.EnjoyVisitMessage, text.TermsTitle, text.TermsIntro,
		text.OnlinePurchaseTitle, text.OnlinePurchaseContent,
		text.PurchaseConfirmTitle, text.PurchaseConfirmContent,
		text.RefundPolicyTitle, text.RefundPolicyContent,
		text.DateChangePolicyTitle, text.DateChangePolicyContent,
		text.LiabilityTitle, text.LiabilityContent))

	// Order summary section
	contentBuilder.WriteString(fmt.Sprintf(`
    <div>
        <div class="order-summary">
            <div class="section-title">
                <h3>%s</h3>
            </div>
            <table class="order-table">
                <thead>
                    <tr>
                        <th>%s</th>
                        <th>%s</th>
                        <th>%s</th>
                        <th>%s</th>
                    </tr>
                </thead>
                <tbody>
    `, text.OrderSummaryTitle, text.ItemsHeader, text.QuantityHeader, text.PriceHeader, text.TotalHeader))

	var subtotal float64
	for _, item := range orderItems {
		total := item.Price * float64(item.Quantity)
		subtotal += total

		contentBuilder.WriteString(`
                    <tr>
                        <td>` + item.Description + `<br><span class="item-date">` + item.EntryDate + `</span></td>
                        <td>` + strconv.Itoa(item.Quantity) + `</td>
                        <td>MYR ` + strconv.FormatFloat(item.Price, 'f', 2, 64) + `</td>
                        <td>MYR ` + fmt.Sprintf("%.2f", total) + `</td>
                    </tr>
        `)
	}

	contentBuilder.WriteString(fmt.Sprintf(`
                </tbody>
                <tfoot>
                    <tr>
                        <td colspan="3">%s</td>
                        <td>MYR %.2f</td>
                    </tr>
                    <tr>
                        <td colspan="3">%s</td>
                        <td>MYR %.2f</td>
                    </tr>
                </tfoot>
            </table>
        </div>
    `, text.SubtotalLabel, subtotal, text.TotalWithGSTLabel, subtotal))

	// Complete HTML email body
	body := fmt.Sprintf(`
<html>
<head>
    <style>
        /* Same CSS as before - keeping it unchanged for brevity */
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; color: #333; background-color: #f9f9f9; }
        .container { max-width: 650px; margin: 0 auto; background-color: #ffffff; border-radius: 5px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { background-color: #D5C58A; color: #000000; padding: 15px; border-radius: 5px 5px 0 0; text-align: center; }
        .order-info-section { background-color: #f8f8f8; border-radius: 5px; padding: 20px; }
        .info-label { color: #666; font-size: 14px; margin-bottom: 5px; }
        .info-value { font-weight: bold; font-size: 16px; color: #333; }
        .redeem-title h4 { font-size: 16px; }
        .redeem-title p { margin: 0; font-size: 14px; color: #666; }
        .order-summary { background-color: #f9f9f9; border-radius: 5px; padding: 15px; margin: 15px 0; }
        .order-table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        .order-table th { background-color: #f0f0f0; text-align: left; padding: 8px; font-size: 14px; border-bottom: 1px solid #ddd; }
        .order-table td { padding: 8px; border-bottom: 1px solid #eee; font-size: 14px; }
        .order-table tfoot td { font-weight: bold; border-top: 2px solid #ddd; }
        .terms-section { margin-top: 30px; padding: 15px; background-color: #f9f9f9; border-radius: 5px; font-size: 12px; }
        .terms-section h4 { margin-top: 0; border-bottom: 1px solid #ddd; padding-bottom: 5px; }
        .footer { margin-top: 20px; padding: 15px 0; text-align: center; font-size: 12px; color: #888; border-top: 1px solid #eee; }
        .section-title { text-align: center; }
        .section-title h3 { margin: 0; color: #333; font-size: 18px; }
        .item-date { font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="text-transform: uppercase;font-size: 32px;">%s</h1>
        	<div class="company-info">
				<p>%s</p>
			</div>
        </div>
        
        %s
        
        <div class="footer">
            <p>%s</p>
            <p>%s %s | %s</p>
            <p>%s</p>
        </div>
    </div>
</body>
</html>
`, orderOverview.TicketGroup, address, contentBuilder.String(),
		text.AutomaticMessage, text.ContactUs, contactNo, email, text.Copyright)

	return subject, body, qrAttachments, nil
}

func generateQRCodeBytes(content string) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("QR content cannot be empty")
	}

	// Generate QR code
	qrCode, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Scale QR code
	scaledQR, err := barcode.Scale(qrCode, 150, 150)
	if err != nil {
		return nil, fmt.Errorf("failed to scale QR code: %w", err)
	}

	// Encode to PNG
	var qrBuffer bytes.Buffer
	err = png.Encode(&qrBuffer, scaledQR)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	return qrBuffer.Bytes(), nil
}
