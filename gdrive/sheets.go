package gdrive

import (
	"context"
	"disaster/model"
	"fmt"
	"log"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GetCategories retrieves all categories from the Google Sheet
func GetCategories(ctx context.Context) ([]model.Category, error) {
	spreadsheetID := "1DX0_eUz1QRe0xWdnVAhIOYYsp7lpOUsmMg74QIjwDac"
	readRange := "Sheet1!C3:E" // Updated to include version column E

	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		log.Printf("Unable to retrieve Sheets client: %v", err)
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from sheet: %v", err)
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// Create a map to store unique categories
	categoryMap := make(map[string]bool)
	var categories []model.Category

	// Iterate through the rows
	for _, row := range resp.Values {
		if len(row) >= 3 { // Make sure we have enough columns
			categoryName := row[0].(string) // Category is in the 1st column since we're get C through E
			versionStr := row[2].(string)   // Version is in the 3rd column
			version, err := strconv.ParseFloat(versionStr, 32)
			if err != nil {
				log.Printf("Warning: invalid version number %q for category %q: %v", versionStr, categoryName, err)
				continue
			}
			if float32(version) > 1.1 {
				continue
			}
			if categoryName != "" && !categoryMap[categoryName] {
				categoryMap[categoryName] = true
				categories = append(categories, model.Category{Name: categoryName})
			}
		}
	}

	return categories, nil
}

// GetResourcesByCategory retrieves all resources for a specific category from the Google Sheet
func GetResourcesByCategory(ctx context.Context, category string) ([]model.Resource, error) {
	spreadsheetID := "1DX0_eUz1QRe0xWdnVAhIOYYsp7lpOUsmMg74QIjwDac"
	readRange := "Sheet1!A3:F" // Include all columns

	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		log.Printf("Unable to retrieve Sheets client: %v", err)
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from sheet: %v", err)
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	var resources []model.Resource

	// Iterate through the rows
	for _, row := range resp.Values {
		if len(row) >= 4 { // Make sure we have enough columns
			rowCategory := row[2].(string) // Category is in the third column
			if rowCategory == category {
				resource := model.Resource{
					Name:        row[0].(string), // Name is in the first column
					Description: row[1].(string), // Description is in the second column
					Category:    rowCategory,
					Link:        row[3].(string), // Link is in the fourth column
				}
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ExtractGoogleDocID extracts the Google Doc ID from a URL
func ExtractGoogleDocID(url string) string {
	// Common patterns for Google Sheets URLs:
	// https://docs.google.com/spreadsheets/d/[ID]/edit
	// https://docs.google.com/spreadsheets/d/[ID]/view
	// https://docs.google.com/spreadsheets/d/[ID]

	patterns := []string{
		"spreadsheets/d/",
		"docs.google.com/spreadsheets/d/",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(url, pattern); idx != -1 {
			start := idx + len(pattern)
			end := strings.IndexAny(url[start:], "/?")
			if end == -1 {
				return url[start:]
			}
			return url[start : start+end]
		}
	}
	return ""
}

// TabInfo contains information about a spreadsheet tab
type TabInfo struct {
	Title     string `json:"title"`
	HasConfig bool   `json:"hasConfig"`
}

// GetSpreadsheetInfo retrieves metadata about a spreadsheet including its tabs
func GetSpreadsheetInfo(ctx context.Context, spreadsheetID string) ([]TabInfo, error) {
	// Check if spreadsheet exists in config
	tabConfig, exists := SheetConfig[spreadsheetID]
	if !exists {
		return nil, fmt.Errorf("no configuration found for spreadsheet ID: %s", spreadsheetID)
	}

	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	// Get spreadsheet metadata
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet metadata: %w", err)
	}

	// Build tab info list
	var tabInfos []TabInfo
	for _, sheet := range spreadsheet.Sheets {
		title := sheet.Properties.Title
		_, hasConfig := tabConfig[title]
		tabInfos = append(tabInfos, TabInfo{
			Title:     title,
			HasConfig: hasConfig,
		})
	}

	return tabInfos, nil
}

// GetSheetData retrieves data from a specific tab and range in a Google Sheet
func GetSheetData(ctx context.Context, spreadsheetID, tabName, dataRange string) ([][]interface{}, error) {
	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		log.Printf("Unable to retrieve Sheets client: %v", err)
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	// Format the range with the tab name
	readRange := fmt.Sprintf("%s!%s", tabName, dataRange)
	
	// Get both values and formatting
	resp, err := srv.Spreadsheets.Get(spreadsheetID).Ranges(readRange).IncludeGridData(true).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from sheet: %v", err)
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Sheets) == 0 || len(resp.Sheets[0].Data) == 0 || len(resp.Sheets[0].Data[0].RowData) == 0 {
		return nil, fmt.Errorf("no data found in range %s", readRange)
	}

	// Convert the response to our desired format
	var result [][]interface{}
	rows := resp.Sheets[0].Data[0].RowData
	for _, row := range rows {
		// Skip empty rows
		if len(row.Values) == 0 {
			continue
		}

		// Check if all cells in the row are empty
		hasContent := false
		for _, cell := range row.Values {
			if cell.FormattedValue != "" {
				hasContent = true
				break
			}
		}
		if !hasContent {
			continue
		}

		var rowData []interface{}
		for _, cell := range row.Values {
			// Check for hyperlink
			if cell.Hyperlink != "" {
				rowData = append(rowData, map[string]interface{}{
					"text": cell.FormattedValue,
					"link": cell.Hyperlink,
				})
			} else {
				rowData = append(rowData, cell.FormattedValue)
			}
		}
		result = append(result, rowData)
	}

	return result, nil
}

// GetSheetDataFromConfig retrieves data from a Google Sheet based on the configuration
func GetSheetDataFromConfig(ctx context.Context, spreadsheetID string) (map[string][][]interface{}, error) {
	// Check if spreadsheet exists in config
	tabConfig, exists := SheetConfig[spreadsheetID]
	if !exists {
		return nil, fmt.Errorf("no configuration found for spreadsheet ID: %s", spreadsheetID)
	}

	result := make(map[string][][]interface{})

	// For each tab in the config, fetch the data
	for tabName, config := range tabConfig {
		if configMap, ok := config.(map[string]interface{}); ok {
			if dataRange, ok := configMap["StructuredDataRange"].(string); ok {
				resp, err := GetSheetData(ctx, spreadsheetID, tabName, dataRange)
				if err != nil {
					return nil, fmt.Errorf("failed to get data for tab %s: %w", tabName, err)
				}
				result[tabName] = resp
			}
		}
	}

	return result, nil
}
