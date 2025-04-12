package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"disaster/gdrive"
)

// HandleSheetTabs handles requests for Google Sheet tab information
func HandleSheetTabs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract sheet ID from URL path
	// Expected format: /api/sheet-tabs/{sheetID}
	sheetID := r.URL.Path[len("/api/sheet-tabs/"):]
	if sheetID == "" {
		log.Printf("Sheet ID not provided in request")
		http.Error(w, "Sheet ID not provided", http.StatusBadRequest)
		return
	}

	// Get sheet tabs
	tabs, err := gdrive.GetSpreadsheetInfo(r.Context(), sheetID)
	if err != nil {
		log.Printf("Error getting spreadsheet info for sheet %s: %v", sheetID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tabs); err != nil {
		log.Printf("Error encoding tabs response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
