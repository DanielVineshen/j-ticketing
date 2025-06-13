// File: j-ticketing/internal/core/services/excel_service.go
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
	f.MergeCell(sheetName, "A1", "B1")

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

	//// Summary section
	//if len(reportData.Summary) > 0 {
	//	row += 2
	//	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Summary")
	//	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), headerStyle)
	//	f.MergeCell(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row))
	//
	//	for key, value := range reportData.Summary {
	//		row++
	//		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), key+":")
	//		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), value)
	//		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), dataStyle)
	//	}
	//}

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
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 40)

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
	f.MergeCell(sheetName, "A1", "C1")

	currentRow := 3

	// Handle summary data in fixed order
	if len(dataSet.SummaryData) > 0 {
		for _, item := range dataSet.SummaryData {
			if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), item.Key); err != nil {
				return fmt.Errorf("failed to set key in A%d: %w", currentRow, err)
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), item.Value); err != nil {
				return fmt.Errorf("failed to set value in C%d: %w", currentRow, err)
			}
			currentRow++
		}
		currentRow += 2 // Add spacing after summary
	}

	// Handle table data with explicit column order
	if len(dataSet.TableData) > 0 {
		headers := dataSet.TableHeaders
		if len(headers) == 0 {
			// Fallback to first row keys if no explicit headers
			for key := range dataSet.TableData[0] {
				headers = append(headers, key)
			}
		}

		row := currentRow

		// Write headers in specified order
		for i, header := range headers {
			col := s.indexToColumn(i)
			f.SetCellValue(sheetName, col+strconv.Itoa(row), header)
			f.SetCellStyle(sheetName, col+strconv.Itoa(row), col+strconv.Itoa(row), headerStyle)
		}

		// Write data using header order
		for _, rowData := range dataSet.TableData {
			row++
			for i, header := range headers {
				col := s.indexToColumn(i)
				value := rowData[header]
				f.SetCellValue(sheetName, col+strconv.Itoa(row), value)
				f.SetCellStyle(sheetName, col+strconv.Itoa(row), col+strconv.Itoa(row), dataStyle)
			}
		}

		// Add chart if configured
		if dataSet.ChartConfig != nil {
			if err := s.addChartToSheet(f, sheetName, dataSet.ChartConfig, len(dataSet.TableData), currentRow, len(dataSet.TableHeaders)); err != nil {
				fmt.Printf("Warning: Failed to add chart to sheet %s: %v\n", sheetName, err)
			}
		}
	}

	// Auto-fit columns
	if len(dataSet.TableHeaders) > 0 {
		for i := range dataSet.TableHeaders {
			col := s.indexToColumn(i)
			f.SetColWidth(sheetName, col, col, 20)
		}
	}

	// Add chart if chart config exists
	if dataSet.ChartConfig != nil {
		if err := s.addChartToSheet(f, sheetName, dataSet.ChartConfig, len(dataSet.TableData), currentRow, len(dataSet.TableHeaders)); err != nil {
			// Log error but don't fail the whole process
			fmt.Printf("Warning: Failed to add chart to sheet %s: %v\n", sheetName, err)
		}
	}

	return nil
}

// addChartToSheet adds a chart to the specified sheet
func (s *ExcelService) addChartToSheet(f *excelize.File, sheetName string, chartConfig *dto.ChartConfig, dataRows, startRow, tableHeadersLen int) error {
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

	var series []excelize.ChartSeries

	// Special handling for New vs Returning Visitors line chart
	if chartConfig.ChartType == "line" && strings.Contains(strings.ToLower(chartConfig.SeriesName), "new vs returning") {
		series = []excelize.ChartSeries{
			{
				Name:       fmt.Sprintf("%s!$B$%d", sheetName, startRow),
				Categories: fmt.Sprintf("%s!$A$%d:$A$%d", sheetName, startRow+1, startRow+dataRows),
				Values:     fmt.Sprintf("%s!$B$%d:$B$%d", sheetName, startRow+1, startRow+dataRows),
			},
			{
				Name:       fmt.Sprintf("%s!$C$%d", sheetName, startRow),
				Categories: fmt.Sprintf("%s!$A$%d:$A$%d", sheetName, startRow+1, startRow+dataRows),
				Values:     fmt.Sprintf("%s!$C$%d:$C$%d", sheetName, startRow+1, startRow+dataRows),
			},
		}
	} else {
		series = []excelize.ChartSeries{
			{
				Name:       fmt.Sprintf("%s!$A$%d", sheetName, startRow),
				Categories: fmt.Sprintf("%s!$A$%d:$A$%d", sheetName, startRow+1, startRow+dataRows),
				Values:     fmt.Sprintf("%s!$B$%d:$B$%d", sheetName, startRow+1, startRow+dataRows),
			},
		}
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
		Type:   chartType,
		Series: series,
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
	chartColumn := s.getColumn("A", tableHeadersLen+1) // Give an extra column between the data and graph
	cell := fmt.Sprintf("%s%d", chartColumn, startRow)
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

func (s *ExcelService) indexToColumnLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

func (s *ExcelService) addToColumn(baseColumn string, offset int) string {
	// Convert base column to index
	baseIndex := s.columnLetterToIndex(baseColumn)
	// Add offset
	newIndex := baseIndex + offset
	// Convert back to letter
	return s.indexToColumnLetter(newIndex)
}

func (s *ExcelService) columnLetterToIndex(column string) int {
	result := 0
	for _, char := range column {
		result = result*26 + int(char-'A'+1)
	}
	return result - 1
}

func (s *ExcelService) getColumn(startColumn string, offset int) string {
	// Method 1: If you know startColumn is always single letter
	if len(startColumn) == 1 {
		baseIndex := int(startColumn[0] - 'A')
		newIndex := baseIndex + offset
		return string(rune('A' + newIndex))
	}

	// Method 2: For multi-letter columns, use the helper functions
	return s.addToColumn(startColumn, offset)
}
