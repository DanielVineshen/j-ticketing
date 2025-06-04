package payment

type TransactionResponse struct {
	IDTransaksi     string `json:"id_transaksi"`
	OrderNo         string `json:"order_no"`
	StatusTransaksi string `json:"status_transaksi"`
	StatusMessage   string `json:"status_message"`
	TarikhTransaksi string `json:"tarikh_transaksi"`
	KodBank         string `json:"kod_bank"`
	NamaBank        string `json:"nama_bank"`
	JpMsgToken      string `json:"jp_msg_token"`
}

type TicketItem struct {
	ItemId string `json:"ItemId"`
	Qty    int    `json:"Qty"`
}

type ZooTicketRequest struct {
	TranDate    string       `json:"TranDate"`
	ReferenceNo string       `json:"ReferenceNo"`
	Items       []TicketItem `json:"Items"`
}

type ZooTicketInfo struct {
	TWBID       string `json:"TWBID"`
	ItemId      string `json:"ItemId"`
	EncryptedID string `json:"EncryptedID"`
	AdmitDate   string `json:"AdmitDate"`
	UnitPrice   string `json:"UnitPrice"`
	ItemDesc    string `json:"ItemDesc"`
	ItemDesc2   string `json:"ItemDesc2"`
	ItemDesc3   string `json:"ItemDesc3"`
}

type ZooTicketResponse struct {
	StatusCode    string          `json:"StatusCode"`
	ReceiptNumber string          `json:"ReceiptNumber"`
	Tickets       []ZooTicketInfo `json:"Tickets"`
}
