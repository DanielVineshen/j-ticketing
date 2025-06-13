// File: j-ticketing/internal/core/dto/report/report.go
package report

import (
	"time"
)

// DTOs for API requests/responses
type CreateReportRequest struct {
	Title       string   `json:"title" validate:"required"`
	Type        string   `json:"type" validate:"required,oneof=onsite_visitors sales members online_visitors"`
	DataOptions []string `json:"dataOptions" validate:"required,min=1"` // Will be joined with semicolon
	Frequency   string   `json:"frequency" validate:"required,oneof=one_time daily weekly monthly quarterly annual"`
	EmailTo     string   `json:"emailTo" validate:"required,email"`
	Desc        *string  `json:"desc"`
}

type UpdateReportRequest struct {
	Title       *string  `json:"title"`
	DataOptions []string `json:"dataOptions"`
	Frequency   *string  `json:"frequency" validate:"omitempty,oneof=one_time daily weekly monthly quarterly annual"`
	EmailTo     *string  `json:"emailTo" validate:"omitempty,email"`
	Desc        *string  `json:"desc"`
}

type ReportResponse struct {
	ReportId          uint                       `json:"reportId"`
	Title             string                     `json:"title"`
	Type              string                     `json:"type"`
	DataOptions       []string                   `json:"dataOptions"` // Split from semicolon-separated
	Frequency         string                     `json:"frequency"`
	EmailTo           string                     `json:"emailTo"`
	Desc              *string                    `json:"desc"`
	IsDeleted         bool                       `json:"isDeleted"`
	CreatedAt         time.Time                  `json:"createdAt"`
	UpdatedAt         time.Time                  `json:"updatedAt"`
	ReportAttachments []ReportAttachmentResponse `json:"attachments"`
}

type ReportAttachmentResponse struct {
	ReportAttachmentId uint      `json:"reportAttachmentId"`
	Type               string    `json:"type"`
	AttachmentName     string    `json:"attachmentName"`
	AttachmentPath     string    `json:"attachmentPath"`
	AttachmentSize     int64     `json:"attachmentSize"`
	ContentType        string    `json:"contentType"`
	CreatedAt          time.Time `json:"createdAt"`
}

// OnsiteVisitors data options enum
type OnsiteVisitorsDataOption string

const (
	TotalOnsiteVisitors    OnsiteVisitorsDataOption = "total_onsite_visitors"
	NewVsReturningVisitors OnsiteVisitorsDataOption = "new_vs_returning_visitors"
	AveragePeakDayAnalysis OnsiteVisitorsDataOption = "average_peak_day_analysis"
	VisitorsByAttraction   OnsiteVisitorsDataOption = "visitors_by_attraction"
	VisitorsByAgeGroup     OnsiteVisitorsDataOption = "visitors_by_age_group"
	VisitorsByNationality  OnsiteVisitorsDataOption = "visitors_by_nationality"
)

// ValidOnsiteVisitorsDataOptions returns all valid options for onsite_visitors type
func ValidOnsiteVisitorsDataOptions() []string {
	return []string{
		string(TotalOnsiteVisitors),
		string(NewVsReturningVisitors),
		string(AveragePeakDayAnalysis),
		string(VisitorsByAttraction),
		string(VisitorsByAgeGroup),
		string(VisitorsByNationality),
	}
}

// GenerateReportRequest represents the request to generate a report
type GenerateReportRequest struct {
	ReportId  uint   `json:"reportId" validate:"required"`
	StartDate string `json:"startDate" validate:"required"` // YYYY-MM-DD format
	EndDate   string `json:"endDate" validate:"required"`   // YYYY-MM-DD format
}

// ExcelReportData represents the structure for Excel data
type ExcelReportData struct {
	ReportTitle string                 `json:"reportTitle"`
	StartDate   string                 `json:"startDate"`
	EndDate     string                 `json:"endDate"`
	DataSets    []ExcelDataSet         `json:"dataSets"`
	Summary     map[string]interface{} `json:"summary"`
}

type ExcelDataSet struct {
	Title        string                   `json:"title"`
	Type         string                   `json:"type"` // table, chart
	SummaryData  []OrderedKeyValue        `json:"summaryData"`
	TableHeaders []string                 `json:"tableHeaders,omitempty"`
	TableData    []map[string]interface{} `json:"tableData,omitempty"`
	ChartConfig  *ChartConfig             `json:"chartConfig,omitempty"`
}

type ChartConfig struct {
	ChartType  string                   `json:"chartType"` // bar, pie, line
	XAxis      string                   `json:"xAxis"`
	YAxis      string                   `json:"yAxis"`
	Data       []map[string]interface{} `json:"data"`
	Width      int                      `json:"width"`
	Height     int                      `json:"height"`
	SeriesName string                   `json:"seriesName,omitempty"`
}

type SeriesConfig struct {
	Name   string `json:"name"`
	Column string `json:"column"`
}

type OrderedKeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
