// 0.go init will always be called first
// Put package level stuff here
package main

import (
	"embed"
	"log"
	"net/http"
	"os"
)

var (
	// Create a new router & API
	router *http.ServeMux

	//go:embed all:static/css/*
	css embed.FS

	//go:embed all:static/font/*
	font embed.FS

	//go:embed all:static/js/*
	js embed.FS

	//go:embed all:static/img/*
	img embed.FS

	//go:embed all:static/svg/*
	svg embed.FS

	// Add global logger declaration at package level
	logger *log.Logger
)

func init() {
	// logging
	logger = log.New(os.Stdout, "http: ", log.LstdFlags)

	router = http.NewServeMux()
	log.Printf("Initialize router")

	cssFS := http.FS(css)
	cssHandler := http.FileServer(cssFS)
	router.Handle("GET /static/css/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[len(r.URL.Path)-4:] == ".css" {
			w.Header().Set("Content-Type", "text/css")
		}
		cssHandler.ServeHTTP(w, r)
	}))

	jsFS := http.FS(js)
	jsHandler := http.FileServer(jsFS)
	router.Handle("GET /static/js/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[len(r.URL.Path)-3:] == ".js" {
			w.Header().Set("Content-Type", "application/javascript")
		}
		jsHandler.ServeHTTP(w, r)
	}))

	imgFS := http.FS(img)
	imgHandler := http.FileServer(imgFS)
	router.Handle("GET /static/img/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extToMime := map[string]string{
			".svg":  "image/svg+xml",
			".png":  "image/png",
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".webp": "image/webp",
			// Add more extensions and MIME types as needed
		}

		ext := r.URL.Path[len(r.URL.Path)-4:]
		if mime, ok := extToMime[ext]; ok {
			w.Header().Set("Content-Type", mime)
		}
		imgHandler.ServeHTTP(w, r)
	}))

	svgFS := http.FS(svg)
	svgHandler := http.FileServer(svgFS)
	router.Handle("GET /static/svg/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extToMime := map[string]string{
			".svg":  "image/svg+xml",
			".png":  "image/png",
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".webp": "image/webp",
			// Add more extensions and MIME types as needed
		}

		ext := r.URL.Path[len(r.URL.Path)-4:]
		if mime, ok := extToMime[ext]; ok {
			w.Header().Set("Content-Type", mime)
		}
		svgHandler.ServeHTTP(w, r)
	}))

	setupRoutes(router)
}
