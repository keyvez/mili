package main

import (
	"net/http"

	"disaster/handlers"
)

func setupRoutes(router *http.ServeMux) {
	router.Handle("GET /", http.HandlerFunc(handlers.Index))
	router.Handle("POST /resources", http.HandlerFunc(handlers.HandleResourcesByCategory))
	router.Handle("GET /api/sheet-tabs/", http.HandlerFunc(handlers.HandleSheetTabs))
	router.Handle("GET /api/sheet-data/", http.HandlerFunc(handlers.HandleSheetData))
	router.Handle("POST /api/render/sheet-tabs", http.HandlerFunc(handlers.HandleRenderSheetTabs))

	// Serve static files with cache headers
	fileServer := http.FileServer(http.Dir("static"))
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" {
			w.Header().Set("Cache-Control", "max-age=31536000") // Cache for 1 year
		}
		fileServer.ServeHTTP(w, r)
	})
	router.Handle("GET /static/", http.StripPrefix("/static/", wrappedHandler))
}
