package service

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/report"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelService handles Excel file generation
type ExcelService struct{}

// NewExcelService creates a new Excel service
func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// GenerateExcelReport generates an Excel file from report data
func (s *ExcelService) GenerateExcelReport(reportData *dto.ExcelReportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Create overview sheet
	if err := s.createOverviewSheet(f, reportData); err != nil {
		return nil, fmt.Errorf("failed to create overview sheet: %w", err)
	}

	for i, dataSet := range reportData.DataSets {
		// Create a more concise sheet name
		baseName := s.sanitizeSheetName(dataSet.Title)
		sheetName := fmt.Sprintf("Data_%d_%s", i+1, baseName)

		// Double-check length after adding prefix
		if len(sheetName) > 31 {
			// If still too long, use a shorter format
			shortName := s.createShortSheetName(dataSet.Title, i+1)
			sheetName = shortName
		}

		if err := s.createDataSheet(f, sheetName, &dataSet); err != nil {
			return nil, fmt.Errorf("failed to create data sheet %s: %w", sheetName, err)
		}
	}

	// Create charts sheet if there are chart configs
	//if s.hasChartData(reportData) {
	//	if err := s.createChartsSheet(f, reportData); err != nil {
	//		return nil, fmt.Errorf("failed to create charts sheet: %w", err)
	//	}
	//}

	// Delete the default Sheet1 if it exists and is empty
	if sheetIndex, err := f.GetSheetIndex("Sheet1"); err == nil && sheetIndex >= 0 {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			// Log but don't fail if we can't delete the default sheet
		}
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// createOverviewSheet creates the overview/summary sheet
func (s *ExcelService) createOverviewSheet(f *excelize.File, reportData *dto.ExcelReportData) error {
	sheetName := "Overview"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// Title and header styling
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#1f4e79"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#ffffff"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472c4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})

	// Set title
	f.SetCellValue(sheetName, "A1", reportData.ReportTitle)
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.MergeCell(sheetName, "A1", "D1")

	// Report info
	row := 3
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Report Information")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), headerStyle)
	f.MergeCell(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row))

	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Start Date:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), reportData.StartDate)
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)

	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "End Date:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), reportData.EndDate)
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)

	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Generated At:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), time.Now().Format("2006-01-02 15:04:05"))
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)

	// Summary section
	if len(reportData.Summary) > 0 {
		row += 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Summary")
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), headerStyle)
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row))

		for key, value := range reportData.Summary {
			row++
			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), key+":")
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), value)
			f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)
		}
	}

	// Data sheets index
	if len(reportData.DataSets) > 0 {
		row += 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Data Sheets")
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), headerStyle)
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row))

		for i, dataSet := range reportData.DataSets {
			row++
			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), fmt.Sprintf("Sheet %d:", i+1))
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), dataSet.Title)
			f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)
		}
	}

	// Auto-fit columns
	f.SetColWidth(sheetName, "A", "B", 20)
	f.SetColWidth(sheetName, "C", "D", 15)

	return nil
}

// createDataSheet creates a sheet for each dataset
func (s *ExcelService) createDataSheet(f *excelize.File, sheetName string, dataSet *dto.ExcelDataSet) error {
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#1f4e79"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#ffffff"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472c4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})

	// Set title
	f.SetCellValue(sheetName, "A1", dataSet.Title)
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	if len(dataSet.TableData) == 0 {
		f.SetCellValue(sheetName, "A3", "No data available")
		return nil
	}

	// Get column headers from first row
	headers := make([]string, 0)
	for key := range dataSet.TableData[0] {
		headers = append(headers, key)
	}

	// Write headers
	row := 3
	for i, header := range headers {
		col := s.indexToColumn(i)
		f.SetCellValue(sheetName, col+strconv.Itoa(row), header)
		f.SetCellStyle(sheetName, col+strconv.Itoa(row), col+strconv.Itoa(row), headerStyle)
	}

	// Write data
	for _, rowData := range dataSet.TableData {
		row++
		for i, header := range headers {
			col := s.indexToColumn(i)
			value := rowData[header]
			f.SetCellValue(sheetName, col+strconv.Itoa(row), value)
			f.SetCellStyle(sheetName, col+strconv.Itoa(row), col+strconv.Itoa(row), dataStyle)
		}
	}

	// Auto-fit columns
	for i := range headers {
		col := s.indexToColumn(i)
		f.SetColWidth(sheetName, col, col, 15)
	}

	// Add chart if chart config exists
	if dataSet.ChartConfig != nil {
		if err := s.addChartToSheet(f, sheetName, dataSet.ChartConfig, len(dataSet.TableData), row+2); err != nil {
			// Log error but don't fail the whole process
			fmt.Printf("Warning: Failed to add chart to sheet %s: %v\n", sheetName, err)
		}
	}

	return nil
}

// createChartsSheet creates a dedicated charts sheet
//func (s *ExcelService) createChartsSheet(f *excelize.File, reportData *dto.ExcelReportData) error {
//	sheetName := "Charts"
//	_, err := f.NewSheet(sheetName)
//	if err != nil {
//		return err
//	}
//
//	titleStyle, _ := f.NewStyle(&excelize.Style{
//		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#1f4e79"},
//		Alignment: &excelize.Alignment{Horizontal: "center"},
//	})
//
//	f.SetCellValue(sheetName, "A1", "Data Visualizations")
//	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
//
//	row := 3
//	for i, dataSet := range reportData.DataSets {
//		if dataSet.ChartConfig != nil {
//			// Create a data table for the chart
//			headers := []string{dataSet.ChartConfig.XAxis, dataSet.ChartConfig.YAxis}
//
//			// Write headers
//			for j, header := range headers {
//				col := s.indexToColumn(j)
//				f.SetCellValue(sheetName, col+strconv.Itoa(row), header)
//			}
//
//			// Write chart data
//			for _, chartRow := range dataSet.ChartConfig.Data {
//				row++
//				f.SetCellValue(sheetName, "A"+strconv.Itoa(row), chartRow[dataSet.ChartConfig.XAxis])
//				f.SetCellValue(sheetName, "B"+strconv.Itoa(row), chartRow[dataSet.ChartConfig.YAxis])
//			}
//
//			// Add chart
//			if err := s.addChartToSheet(f, sheetName, dataSet.ChartConfig, len(dataSet.ChartConfig.Data), row+2); err != nil {
//				fmt.Printf("Warning: Failed to add chart %d: %v\n", i, err)
//			}
//
//			row += 15 // Space for next chart
//		}
//	}
//
//	return nil
//}

// addChartToSheet adds a chart to the specified sheet
func (s *ExcelService) addChartToSheet(f *excelize.File, sheetName string, chartConfig *dto.ChartConfig, dataRows, startRow int) error {
	var chartType excelize.ChartType
	switch chartConfig.ChartType {
	case "bar":
		chartType = excelize.Col
	case "pie":
		chartType = excelize.Pie
	case "line":
		chartType = excelize.Line
	default:
		chartType = excelize.Col
	}

	// Use SeriesName if provided, otherwise use default
	seriesName := chartConfig.XAxis + " vs " + chartConfig.YAxis
	if chartConfig.SeriesName != "" {
		seriesName = chartConfig.SeriesName
	}

	// Calculate dynamic height based on table
	// Table structure: 1 title row + 1 empty row + 1 header row + dataRows
	totalTableRows := 1 + 1 + 1 + dataRows // 3 + dataRows

	// Excel row height is typically 15 pixels per row
	// Add some padding for a nice look
	dynamicHeight := (totalTableRows * 15) + 20 // 20px padding

	// Set minimum and maximum bounds
	minHeight := 200
	maxHeight := 800
	if dynamicHeight < minHeight {
		dynamicHeight = minHeight
	}
	if dynamicHeight > maxHeight {
		dynamicHeight = maxHeight
	}

	// Set default dimensions
	//width := 480            // Excel default width
	height := dynamicHeight // Use calculated height
	width := height * 2

	// Override with ChartConfig values if provided
	if chartConfig.Width > 0 {
		width = chartConfig.Width
	}
	if chartConfig.Height > 0 {
		height = chartConfig.Height // Manual override takes precedence
	}

	chart := &excelize.Chart{
		Type: chartType,
		Series: []excelize.ChartSeries{
			{
				Name:       seriesName,
				Categories: fmt.Sprintf("%s!$A$4:$A$%d", sheetName, 3+dataRows),
				Values:     fmt.Sprintf("%s!$B$4:$B$%d", sheetName, 3+dataRows),
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     chartConfig.ChartType == "pie",
			ShowSerName:     false,
			ShowVal:         true,
		},
		ShowBlanksAs: "zero",
		Dimension: excelize.ChartDimension{
			Width:  uint(width),
			Height: uint(height), // Dynamic height
		},
	}

	// Place chart next to table
	//cell := fmt.Sprintf("D%d", startRow)
	cell := fmt.Sprintf("E3") // Fixed position next to table
	if err := f.AddChart(sheetName, cell, chart); err != nil {
		return err
	}

	return nil
}

// Helper functions
func (s *ExcelService) indexToColumn(index int) string {
	column := ""
	for index >= 0 {
		column = string(rune('A'+index%26)) + column
		index = index/26 - 1
	}
	return column
}

func (s *ExcelService) createShortSheetName(title string, index int) string {
	// Create abbreviations for common report titles
	abbreviations := map[string]string{
		"Total Onsite Visitors":     "Total_Visitors",
		"New vs Returning Visitors": "New_vs_Return",
		"Average Peak Day Analysis": "Peak_Day_Analysis",
		"Visitors by Attraction":    "By_Attraction",
		"Visitors by Age Group":     "By_Age_Group",
		"Visitors by Nationality":   "By_Nationality",
		"New Vs Returning Visitors": "New_vs_Return",
		"Visitors By Attraction":    "By_Attraction",
		"Visitors By Age Group":     "By_Age_Group",
		"Visitors By Nationality":   "By_Nationality",
	}

	// Check if we have a specific abbreviation
	if abbrev, exists := abbreviations[title]; exists {
		sheetName := fmt.Sprintf("D%d_%s", index, abbrev)
		if len(sheetName) <= 31 {
			return sheetName
		}
	}

	// Fallback: create a very short name
	sanitized := s.sanitizeSheetName(title)
	maxTitleLength := 31 - len(fmt.Sprintf("D%d_", index))

	if len(sanitized) > maxTitleLength {
		sanitized = sanitized[:maxTitleLength]
		sanitized = strings.TrimRight(sanitized, "_")
	}

	return fmt.Sprintf("D%d_%s", index, sanitized)
}

func (s *ExcelService) sanitizeSheetName(name string) string {
	// Excel sheet names can't contain certain characters
	invalidChars := []string{"\\", "/", "*", "?", ":", "[", "]"}
	result := name
	for _, invalidChar := range invalidChars {
		result = strings.ReplaceAll(result, invalidChar, "_")
	}

	// Remove common words to shorten the name
	commonWords := []string{" vs ", " and ", " by ", " for ", " the ", " of ", " in ", " on ", " at "}
	for _, word := range commonWords {
		result = strings.ReplaceAll(result, word, "_")
	}

	// Replace multiple spaces/underscores with single underscore
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	result = strings.ReplaceAll(result, " ", "_")

	// Trim underscores from start and end
	result = strings.Trim(result, "_")

	// Limit length to 31 characters (Excel limit)
	if len(result) > 31 {
		result = result[:31]
		// Remove trailing underscore if present after truncation
		result = strings.TrimRight(result, "_")
	}

	// Ensure we have a valid name (not empty)
	if result == "" {
		result = "Sheet"
	}

	return result
}

func (s *ExcelService) hasChartData(reportData *dto.ExcelReportData) bool {
	for _, dataSet := range reportData.DataSets {
		if dataSet.ChartConfig != nil {
			return true
		}
	}
	return false
}
