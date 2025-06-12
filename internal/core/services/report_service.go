package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dto "j-ticketing/internal/core/dto/report"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/email"
)

// ReportService handles report operations
type ReportService struct {
	reportRepo              repositories.ReportRepository
	reportAttachmentRepo    repositories.ReportAttachmentRepository
	orderTicketGroupService *OrderTicketGroupService
	excelService            *ExcelService
	emailService            email.EmailService
}

// NewReportService creates a new report service
func NewReportService(
	reportRepo repositories.ReportRepository,
	reportAttachmentRepo repositories.ReportAttachmentRepository,
	orderTicketGroupService *OrderTicketGroupService,
	excelService *ExcelService,
	emailService email.EmailService,
) *ReportService {
	return &ReportService{
		reportRepo:              reportRepo,
		reportAttachmentRepo:    reportAttachmentRepo,
		orderTicketGroupService: orderTicketGroupService,
		excelService:            excelService,
		emailService:            emailService,
	}
}

// CreateReport creates a new report configuration
func (s *ReportService) CreateReport(req *dto.CreateReportRequest) (*dto.ReportResponse, error) {
	// Validate data options based on report type
	if err := s.validateDataOptions(req.Type, req.DataOptions); err != nil {
		return nil, err
	}

	// Join data options with semicolon
	dataOptionsStr := strings.Join(req.DataOptions, ";")

	report := &models.Report{
		Title:       req.Title,
		Type:        req.Type,
		DataOptions: dataOptionsStr,
		Frequency:   req.Frequency,
		EmailTo:     req.EmailTo,
		Desc:        req.Desc,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return s.convertToReportResponse(report), nil
}

// GenerateReport generates a report based on configuration and date range
func (s *ReportService) GenerateReport(req *dto.GenerateReportRequest) (*models.ReportAttachment, error) {
	// Get report configuration
	report, err := s.reportRepo.FindByID(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("failed to find report: %w", err)
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	// Parse data options
	dataOptions := strings.Split(report.DataOptions, ";")

	// Generate report data based on type
	var reportData *dto.ExcelReportData
	switch report.Type {
	case "onsite_visitors":
		reportData, err = s.generateOnsiteVisitorsReport(report, dataOptions, req.StartDate, req.EndDate)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", report.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Generate Excel file
	excelFile, err := s.excelService.GenerateExcelReport(reportData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate excel file: %w", err)
	}

	// Save file and create attachment record
	attachment, err := s.saveReportAttachment(report, excelFile, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to save report attachment: %w", err)
	}

	// Send email if configured
	if report.EmailTo != "" {
		go s.sendReportEmail(report, attachment)
	}

	return attachment, nil
}

// generateOnsiteVisitorsReport generates data for onsite visitors report
func (s *ReportService) generateOnsiteVisitorsReport(report *models.Report, dataOptions []string, startDate, endDate string) (*dto.ExcelReportData, error) {
	reportData := &dto.ExcelReportData{
		ReportTitle: report.Title,
		StartDate:   startDate,
		EndDate:     endDate,
		DataSets:    []dto.ExcelDataSet{},
		Summary:     make(map[string]interface{}),
	}

	for _, option := range dataOptions {
		switch strings.TrimSpace(option) {
		case string(dto.TotalOnsiteVisitors):
			if err := s.addTotalOnsiteVisitorsData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		case string(dto.NewVsReturningVisitors):
			if err := s.addNewVsReturningVisitorsData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		case string(dto.AveragePeakDayAnalysis):
			if err := s.addAveragePeakDayAnalysisData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		case string(dto.VisitorsByAttraction):
			if err := s.addVisitorsByAttractionData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		case string(dto.VisitorsByAgeGroup):
			if err := s.addVisitorsByAgeGroupData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		case string(dto.VisitorsByNationality):
			if err := s.addVisitorsByNationalityData(reportData, startDate, endDate); err != nil {
				return nil, err
			}
		}
	}

	return reportData, nil
}

func (s *ReportService) addTotalOnsiteVisitorsData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	// Get data from service - returns *OrderDateRangeResponse
	data, err := s.orderTicketGroupService.GetTotalOnsiteVisitorsWithinRange(startDate, endDate)
	if err != nil {
		return err
	}

	summaryData := []dto.OrderedKeyValue{
		{Key: "Start Date", Value: data.StartDate},
		{Key: "End Date", Value: data.EndDate},
		{Key: "Current Total Records", Value: data.CurrentTotalRecords},
		{Key: "Past Total Records", Value: data.PastTotalRecords},
		{Key: "Diff From Past Period (%)", Value: data.DiffFromPastPeriod},
	}

	tableHeaders := []string{"Date", "Count", "Day"}

	// Daily data table
	dailyTableData := make([]map[string]interface{}, 0, len(data.DailyData))
	for _, daily := range data.DailyData {
		// Parse date to get day of week for better readability
		parsedDate, _ := time.Parse("2006-01-02", daily.Date)
		dayOfWeek := parsedDate.Weekday().String()

		dailyTableData = append(dailyTableData, map[string]interface{}{
			"Date":  daily.Date,
			"Count": daily.Count,
			"Day":   dayOfWeek,
		})
	}

	// Chart data - ensure it uses the exact field names you want
	chartData := make([]map[string]interface{}, 0, len(data.DailyData))
	for _, daily := range data.DailyData {
		chartData = append(chartData, map[string]interface{}{
			"Date":  daily.Date,  // X-axis
			"Count": daily.Count, // Y-axis
		})
	}

	dailyDataSet := dto.ExcelDataSet{
		Title:        "Daily Visitor Data",
		Type:         "table",
		SummaryData:  summaryData,
		TableHeaders: tableHeaders,
		TableData:    dailyTableData,
		ChartConfig: &dto.ChartConfig{
			ChartType:  "bar",
			XAxis:      "Date",  // Must match the key in chartData
			YAxis:      "Count", // Must match the key in chartData
			Data:       chartData,
			SeriesName: "Daily Visitors",
		},
	}

	// Add dataset to report
	reportData.DataSets = append(reportData.DataSets, dailyDataSet)

	// Add basic summary for report metadata
	reportData.Summary["totalVisitors"] = data.CurrentTotalRecords
	reportData.Summary["startDate"] = data.StartDate
	reportData.Summary["endDate"] = data.EndDate

	return nil
}

func (s *ReportService) addNewVsReturningVisitorsData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	data, err := s.orderTicketGroupService.GetNewVsReturningVisitors(startDate, endDate)
	if err != nil {
		return err
	}

	summaryData := []dto.OrderedKeyValue{
		{Key: "Start Date", Value: data.StartDate},
		{Key: "End Date", Value: data.EndDate},
		{Key: "New Total Records", Value: data.CurrentNewTotalRecords},
		{Key: "Returning Total Records", Value: data.CurrentReturningTotalRecords},
		{Key: "Diff From Past Period (%)", Value: data.DiffNewFromPastPeriod},
	}

	tableHeaders := []string{"Date", "New Visitors", "Returning Visitors", "Total Visitors"}

	// Create table data with all three columns
	tableData := make([]map[string]interface{}, 0)
	chartData := make([]map[string]interface{}, 0)

	// Process daily data to show trends over time
	for _, daily := range data.DailyData { // Assuming you have daily breakdown
		tableData = append(tableData, map[string]interface{}{
			"Date":               daily.Date,
			"New Visitors":       daily.NewCount,
			"Returning Visitors": daily.ReturningCount,
			"Total Visitors":     daily.NewCount + daily.ReturningCount,
		})

		// Chart data needs both series
		chartData = append(chartData, map[string]interface{}{
			"Date":               daily.Date,
			"New Visitors":       daily.NewCount,
			"Returning Visitors": daily.ReturningCount,
		})
	}

	dataSet := dto.ExcelDataSet{
		Title:        "New vs Returning Visitors Trends",
		Type:         "table",
		SummaryData:  summaryData,
		TableData:    tableData,
		TableHeaders: tableHeaders,
		ChartConfig: &dto.ChartConfig{
			ChartType:  "line",
			XAxis:      "Date",
			YAxis:      "Visitors",
			Data:       chartData,
			SeriesName: "New vs Returning Visitors",
		},
	}

	reportData.DataSets = append(reportData.DataSets, dataSet)
	return nil
}

func (s *ReportService) addAveragePeakDayAnalysisData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	data, err := s.orderTicketGroupService.GetAveragePeakDayAnalysis(startDate, endDate)
	if err != nil {
		return err
	}

	summaryData := []dto.OrderedKeyValue{
		{Key: "Start Date", Value: data.StartDate},
		{Key: "End Date", Value: data.EndDate},
		{Key: "Peak Day", Value: data.PeakDay},
		{Key: "Peak Visitors", Value: data.PeakDayCount},
	}

	tableHeaders := []string{"Day of Week", "Total Count", "Average"}

	// Convert weekly data to table
	tableData := make([]map[string]interface{}, 0)
	chartData := make([]map[string]interface{}, 0)

	for _, weekData := range data.WeeklyData {
		tableData = append(tableData, map[string]interface{}{
			"Day of Week": weekData.DayOfWeek,
			"Total Count": weekData.Count,
			"Average":     weekData.Average,
		})

		chartData = append(chartData, map[string]interface{}{
			"Day":   weekData.DayOfWeek,
			"Count": weekData.Count,
		})
	}

	dataSet := dto.ExcelDataSet{
		Title:        "Peak Day Analysis",
		Type:         "table",
		SummaryData:  summaryData,
		TableData:    tableData,
		TableHeaders: tableHeaders,
		ChartConfig: &dto.ChartConfig{
			ChartType: "bar",
			XAxis:     "Day",
			YAxis:     "Count",
			Data:      chartData,
		},
	}

	reportData.DataSets = append(reportData.DataSets, dataSet)
	reportData.Summary["peakDay"] = data.PeakDay
	reportData.Summary["peakDayCount"] = data.PeakDayCount
	return nil
}

func (s *ReportService) addVisitorsByAttractionData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	data, err := s.orderTicketGroupService.GetVisitorsByAttraction(startDate, endDate)
	if err != nil {
		return err
	}

	tableHeaders := []string{"Attraction", "Total Visitors", "Percentage"}

	// Convert to table data
	tableData := make([]map[string]interface{}, 0)
	chartData := make([]map[string]interface{}, 0)

	for _, attraction := range data.AttractionData {
		tableData = append(tableData, map[string]interface{}{
			"Attraction":     attraction.TicketGroupName,
			"Total Visitors": attraction.TotalVisitors,
			"Percentage":     fmt.Sprintf("%.2f%%", attraction.Percentage),
		})

		if attraction.TotalVisitors > 0 { // Only include in chart if has visitors
			chartData = append(chartData, map[string]interface{}{
				"Attraction": attraction.TicketGroupName,
				"Visitors":   attraction.TotalVisitors,
			})
		}
	}

	dataSet := dto.ExcelDataSet{
		Title:        "Visitors by Attraction",
		Type:         "table",
		TableData:    tableData,
		TableHeaders: tableHeaders,
		ChartConfig: &dto.ChartConfig{
			ChartType: "pie",
			XAxis:     "Attraction",
			YAxis:     "Visitors",
			Data:      chartData,
		},
	}

	reportData.DataSets = append(reportData.DataSets, dataSet)
	return nil
}

func (s *ReportService) addVisitorsByAgeGroupData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	data, err := s.orderTicketGroupService.GetVisitorsByAgeGroup(startDate, endDate)
	if err != nil {
		return err
	}

	tableHeaders := []string{"Age Group", "Total Visitors", "Percentage"}

	// Convert VisitorsByAgeGroupResponse to table data
	tableData := make([]map[string]interface{}, 0)
	chartData := make([]map[string]interface{}, 0)

	// Process the actual VisitorsByAgeGroupResponse structure
	for _, ageGroupData := range data.AgeGroupData {
		tableData = append(tableData, map[string]interface{}{
			"Age Group":      ageGroupData.AgeGroup,
			"Total Visitors": ageGroupData.TotalVisitors,
			"Percentage":     fmt.Sprintf("%.2f%%", ageGroupData.Percentage),
		})

		// Only include in chart if there are visitors
		if ageGroupData.TotalVisitors > 0 {
			chartData = append(chartData, map[string]interface{}{
				"AgeGroup": ageGroupData.AgeGroup,
				"Visitors": ageGroupData.TotalVisitors,
			})
		}
	}

	dataSet := dto.ExcelDataSet{
		Title:        "Visitors by Age Group",
		Type:         "table",
		TableData:    tableData,
		TableHeaders: tableHeaders,
		ChartConfig: &dto.ChartConfig{
			ChartType: "pie",
			XAxis:     "AgeGroup",
			YAxis:     "Visitors",
			Data:      chartData,
		},
	}

	reportData.DataSets = append(reportData.DataSets, dataSet)
	reportData.Summary["ageGroupBreakdown"] = fmt.Sprintf("Total: %d visitors across %d age groups", data.TotalVisitors, len(data.AgeGroupData))
	return nil
}

func (s *ReportService) addVisitorsByNationalityData(reportData *dto.ExcelReportData, startDate, endDate string) error {
	data, err := s.orderTicketGroupService.GetVisitorsByNationality(startDate, endDate)
	if err != nil {
		return err
	}

	tableHeaders := []string{"Nationality", "Total Visitors", "Percentage"}

	// Convert VisitorsByNationalityResponse to table data
	tableData := make([]map[string]interface{}, 0)
	chartData := make([]map[string]interface{}, 0)

	// Process the actual VisitorsByNationalityResponse structure
	for _, nationalityData := range data.NationalityData {
		tableData = append(tableData, map[string]interface{}{
			"Nationality":    nationalityData.Nationality,
			"Total Visitors": nationalityData.TotalVisitors,
			"Percentage":     fmt.Sprintf("%.2f%%", nationalityData.Percentage),
		})

		// Only include in chart if there are visitors
		if nationalityData.TotalVisitors > 0 {
			chartData = append(chartData, map[string]interface{}{
				"Nationality": nationalityData.Nationality,
				"Visitors":    nationalityData.TotalVisitors,
			})
		}
	}

	dataSet := dto.ExcelDataSet{
		Title:        "Visitors by Nationality",
		Type:         "table",
		TableData:    tableData,
		TableHeaders: tableHeaders,
		ChartConfig: &dto.ChartConfig{
			ChartType: "bar",
			XAxis:     "Nationality",
			YAxis:     "Visitors",
			Data:      chartData,
		},
	}

	reportData.DataSets = append(reportData.DataSets, dataSet)

	// Calculate local vs international breakdown for summary
	var localVisitors, internationalVisitors int
	var localPercentage, internationalPercentage float64

	for _, nationalityData := range data.NationalityData {
		if nationalityData.Nationality == "Local" {
			localVisitors = nationalityData.TotalVisitors
			localPercentage = nationalityData.Percentage
		} else if nationalityData.Nationality == "International" {
			internationalVisitors = nationalityData.TotalVisitors
			internationalPercentage = nationalityData.Percentage
		}
	}

	reportData.Summary["nationalityBreakdown"] = fmt.Sprintf("Local: %d (%.1f%%), International: %d (%.1f%%)",
		localVisitors, localPercentage,
		internationalVisitors, internationalPercentage)

	return nil
}

// Helper methods
func (s *ReportService) validateDataOptions(reportType string, dataOptions []string) error {
	var validOptions []string

	switch reportType {
	case "onsite_visitors":
		validOptions = dto.ValidOnsiteVisitorsDataOptions()
	default:
		return fmt.Errorf("unsupported report type: %s", reportType)
	}

	// Check if all provided options are valid
	for _, option := range dataOptions {
		valid := false
		for _, validOption := range validOptions {
			if option == validOption {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid data option '%s' for report type '%s'", option, reportType)
		}
	}

	return nil
}

func (s *ReportService) saveReportAttachment(report *models.Report, excelFile []byte, startDate, endDate string) (*models.ReportAttachment, error) {
	// Generate unique filename
	timestamp := time.Now().Format("20060102_150405")
	randomStr := s.generateRandomString(8)
	filename := fmt.Sprintf("%s_%s_%s_to_%s_%s.xlsx",
		strings.ReplaceAll(report.Title, " ", "_"),
		timestamp,
		startDate,
		endDate,
		randomStr,
	)

	// Get storage path
	storagePath := os.Getenv("REPORT_STORAGE_PATH")
	if storagePath == "" {
		return nil, fmt.Errorf("REPORT_STORAGE_PATH environment variable not set")
	}

	// Save file
	filePath := filepath.Join(storagePath, filename)
	if err := os.WriteFile(filePath, excelFile, 0644); err != nil {
		return nil, fmt.Errorf("failed to save excel file: %w", err)
	}

	// Create attachment record
	attachment := &models.ReportAttachment{
		ReportId:        report.ReportId,
		Type:            "excel",
		EmailTo:         report.EmailTo,
		AttachmentName:  filename,
		AttachmentPath:  filePath,
		AttachmentSize:  int64(len(excelFile)),
		ContentType:     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		UniqueExtension: randomStr,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.reportAttachmentRepo.Create(attachment); err != nil {
		// Clean up file on database error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	return attachment, nil
}

func (s *ReportService) generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func (s *ReportService) sendReportEmail(report *models.Report, attachment *models.ReportAttachment) {
	// Implementation for sending email with attachment
}

// GetReport retrieves a single report by ID
func (s *ReportService) GetReport(reportId uint) (*dto.ReportResponse, error) {
	report, err := s.reportRepo.FindByID(reportId)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, nil
	}
	return s.convertToReportResponse(report), nil
}

// GetAllReports retrieves all active reports
func (s *ReportService) GetAllReports() ([]dto.ReportResponse, error) {
	reports, err := s.reportRepo.FindActiveReports()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = *s.convertToReportResponse(&report)
	}
	return responses, nil
}

// GetReportsByType retrieves reports by type
func (s *ReportService) GetReportsByType(reportType string) ([]dto.ReportResponse, error) {
	reports, err := s.reportRepo.FindByType(reportType)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = *s.convertToReportResponse(&report)
	}
	return responses, nil
}

// GetReportsByFrequency retrieves reports by frequency
func (s *ReportService) GetReportsByFrequency(frequency string) ([]dto.ReportResponse, error) {
	reports, err := s.reportRepo.FindByFrequency(frequency)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = *s.convertToReportResponse(&report)
	}
	return responses, nil
}

// UpdateReport updates an existing report
func (s *ReportService) UpdateReport(reportId uint, req *dto.UpdateReportRequest) (*dto.ReportResponse, error) {
	report, err := s.reportRepo.FindByID(reportId)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	// Update fields if provided
	if req.Title != nil {
		report.Title = *req.Title
	}
	if req.DataOptions != nil {
		// Validate data options
		if err := s.validateDataOptions(report.Type, req.DataOptions); err != nil {
			return nil, err
		}
		report.DataOptions = strings.Join(req.DataOptions, ";")
	}
	if req.Frequency != nil {
		report.Frequency = *req.Frequency
	}
	if req.EmailTo != nil {
		report.EmailTo = *req.EmailTo
	}
	if req.Desc != nil {
		report.Desc = req.Desc
	}

	report.UpdatedAt = time.Now()

	if err := s.reportRepo.Update(report); err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return s.convertToReportResponse(report), nil
}

// DeleteReport soft deletes a report
func (s *ReportService) DeleteReport(reportId uint) error {
	return s.reportRepo.SoftDelete(reportId)
}

// GetReportAttachment retrieves a single attachment
func (s *ReportService) GetReportAttachment(attachmentId uint) (*models.ReportAttachment, error) {
	return s.reportAttachmentRepo.FindByID(attachmentId)
}

// GetReportAttachments retrieves all attachments for a report
func (s *ReportService) GetReportAttachments(reportId uint) ([]dto.ReportAttachmentResponse, error) {
	attachments, err := s.reportAttachmentRepo.FindByReportID(reportId)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReportAttachmentResponse, len(attachments))
	for i, att := range attachments {
		responses[i] = dto.ReportAttachmentResponse{
			ReportAttachmentId: att.ReportAttachmentId,
			Type:               att.Type,
			AttachmentName:     att.AttachmentName,
			AttachmentPath:     att.AttachmentPath,
			AttachmentSize:     att.AttachmentSize,
			ContentType:        att.ContentType,
			CreatedAt:          att.CreatedAt,
		}
	}
	return responses, nil
}

// PreviewReport generates preview data without saving files
func (s *ReportService) PreviewReport(req *dto.GenerateReportRequest) (*dto.ExcelReportData, error) {
	// Get report configuration
	report, err := s.reportRepo.FindByID(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("failed to find report: %w", err)
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	// Parse data options
	dataOptions := strings.Split(report.DataOptions, ";")

	// Generate preview data based on type
	switch report.Type {
	case "onsite_visitors":
		return s.generateOnsiteVisitorsReport(report, dataOptions, req.StartDate, req.EndDate)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", report.Type)
	}
}

func (s *ReportService) convertToReportResponse(report *models.Report) *dto.ReportResponse {
	dataOptions := strings.Split(report.DataOptions, ";")

	attachments := make([]dto.ReportAttachmentResponse, len(report.ReportAttachments))
	for i, att := range report.ReportAttachments {
		attachments[i] = dto.ReportAttachmentResponse{
			ReportAttachmentId: att.ReportAttachmentId,
			Type:               att.Type,
			AttachmentName:     att.AttachmentName,
			AttachmentPath:     att.AttachmentPath,
			AttachmentSize:     att.AttachmentSize,
			ContentType:        att.ContentType,
			CreatedAt:          att.CreatedAt,
		}
	}

	return &dto.ReportResponse{
		ReportId:          report.ReportId,
		Title:             report.Title,
		Type:              report.Type,
		DataOptions:       dataOptions,
		Frequency:         report.Frequency,
		EmailTo:           report.EmailTo,
		Desc:              report.Desc,
		IsDeleted:         report.IsDeleted,
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
		ReportAttachments: attachments,
	}
}
