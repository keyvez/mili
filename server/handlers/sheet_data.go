package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"disaster/components"
	"disaster/components/sheet_row_cards"
	"disaster/gdrive"
)

// TabsRequest represents the request body for rendering sheet tabs
type TabsRequest struct {
	SheetID string `json:"sheetId"`
}

// HandleRenderSheetTabs handles the request to render sheet tabs
func HandleRenderSheetTabs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sheetID string
	if r.Method == http.MethodGet {
		// Extract sheet ID from URL path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}
		sheetID = pathParts[3]
		if sheetID == "" {
			http.Error(w, "Sheet ID is required", http.StatusBadRequest)
			return
		}
		// URL decode the sheet ID
		var err error
		sheetID, err = url.QueryUnescape(sheetID)
		if err != nil {
			http.Error(w, "Invalid sheet ID", http.StatusBadRequest)
			return
		}
	} else {
		// Parse request body for POST
		var req TabsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		sheetID = req.SheetID
	}

	// Get the sheet tabs
	tabs, err := gdrive.GetSpreadsheetInfo(r.Context(), sheetID)
	if err != nil {
		log.Printf("Error getting sheet tabs: %v", err)
		http.Error(w, "Failed to get sheet tabs", http.StatusInternalServerError)
		return
	}

	// Render the tabs component
	var buf bytes.Buffer
	err = components.SheetTabs(components.SheetTabsProps{
		SheetID: sheetID,
		Tabs:    tabs,
	}).Render(r.Context(), &buf)
	if err != nil {
		log.Printf("Error rendering tabs: %v", err)
		http.Error(w, "Failed to render tabs", http.StatusInternalServerError)
		return
	}

	// Return the rendered HTML directly
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

// HandleSheetData handles the request to get sheet data
func HandleSheetData(w http.ResponseWriter, r *http.Request) {
	// Get sheet ID and tab name from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	sheetID := pathParts[3]
	tabName := pathParts[4]

	// URL decode the sheet ID and tab name
	var err error
	sheetID, err = url.QueryUnescape(sheetID)
	if err != nil {
		http.Error(w, "Invalid sheet ID", http.StatusBadRequest)
		return
	}
	tabName, err = url.QueryUnescape(tabName)
	if err != nil {
		http.Error(w, "Invalid tab name", http.StatusBadRequest)
		return
	}

	// Get the component config for this tab
	tabConfig, exists := gdrive.SheetConfig[sheetID]
	if !exists {
		log.Printf("No config found for sheet: %s", sheetID)
		http.Error(w, "Sheet not configured", http.StatusInternalServerError)
		return
	}

	componentConfig, ok := tabConfig[tabName].(map[string]interface{})
	if !ok {
		log.Printf("Invalid tab configuration for sheet %s, tab %s", sheetID, tabName)
		http.Error(w, "Invalid tab configuration", http.StatusInternalServerError)
		return
	}

	// Get the component type and data range
	componentName, ok := componentConfig["Component"].(string)
	if !ok {
		log.Printf("No component specified in configuration for sheet %s, tab %s", sheetID, tabName)
		http.Error(w, "No component specified in configuration", http.StatusInternalServerError)
		return
	}

	// Get the data from the sheet
	dataRange, ok := componentConfig["StructuredDataRange"].(string)
	if !ok {
		log.Printf("No data range specified in configuration for sheet %s, tab %s", sheetID, tabName)
		http.Error(w, "No data range specified in configuration", http.StatusInternalServerError)
		return
	}

	// Get the card type
	cardType, ok := sheet_row_cards.GetCardType(componentName)
	if !ok {
		log.Printf("Unknown component type: %s", componentName)
		http.Error(w, "Unknown component type", http.StatusInternalServerError)
		return
	}

	// Get the data from the sheet
	data, err := gdrive.GetSheetData(r.Context(), sheetID, tabName, dataRange)
	if err != nil {
		log.Printf("Error getting sheet data: %v", err)
		http.Error(w, "Failed to get sheet data", http.StatusInternalServerError)
		return
	}

	// Get headers and rows
	if len(data) < 1 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	headers := data[0]
	rows := data[1:]

	log.Printf("Headers: %v", headers)

	// Map column names to indices
	colMap := make(map[string]int)
	for j, header := range headers {
		headerStr, ok := header.(string)
		if !ok {
			log.Printf("Warning: header at index %d is not a string: %v", j, header)
			continue
		}
		colMap[headerStr] = j
		log.Printf("Found column: %s at index %d", headerStr, j)
	}

	// Parse rows into structs
	var rowsData []any
	for i, row := range rows {
		log.Printf("Processing row %d: %v", i, row)
		rowData, err := sheet_row_cards.ParseRowData(cardType, row, colMap)
		if err != nil {
			log.Printf("Warning: error parsing row %d: %v", i, err)
			continue
		}
		rowsData = append(rowsData, rowData)
	}

	// Get card renderer for this type
	renderer := cardType.RenderFunc
	if renderer == nil {
		http.Error(w, "No renderer found for card type", http.StatusInternalServerError)
		return
	}

	// Render using the card type's render function
	var buf bytes.Buffer
	err = sheet_row_cards.RowCardContainer(rowsData, renderer).Render(r.Context(), &buf)
	if err != nil {
		http.Error(w, "Failed to render component", http.StatusInternalServerError)
		return
	}

	// Return the rendered HTML
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"html": buf.String(),
	})
}
