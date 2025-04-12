package handlers

import (
	"disaster/components"
	"disaster/gdrive"
	"log"
	"net/http"
)

func HandleResourcesByCategory(w http.ResponseWriter, r *http.Request) {
	// Add debug logging
	log.Printf("Handling resources request for URL: %s", r.URL.String())

	category := r.URL.Query().Get("category")
	if category == "" {
		log.Printf("No category provided")
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching resources for category: %s", category)
	resources, err := gdrive.GetResourcesByCategory(r.Context(), category)
	if err != nil {
		log.Printf("Error fetching resources: %v", err)
		http.Error(w, "Failed to fetch resources", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d resources", len(resources))
	components.ResourcesList(resources).Render(r.Context(), w)
}
