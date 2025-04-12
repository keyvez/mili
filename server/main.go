package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jritsema/gotoolbox"
)

func main() {
	//exit process immediately upon sigterm
	handleSigTerms()

	// tracing
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	middleware := tracing(nextRequestID)(logging(logger)(router))

	port := gotoolbox.GetEnvWithDefault("PORT", "8080")
	logger.Println("listening on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, middleware); err != nil {
		logger.Println("http.ListenAndServe():", err)
		os.Exit(1)
	}
}

func handleSigTerms() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("received SIGTERM, exiting")
		os.Exit(1)
	}()
}
