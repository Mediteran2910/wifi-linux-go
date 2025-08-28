package main

import (
	"log"
	"net/http"
	"time"

	"my-go-app/handlers"
)

func main() {
	// The Go app's sole purpose is to serve the captive portal UI and API.

	// Configure static file serving.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Configure API handlers for Wi-Fi management.
	// http.HandleFunc("/api/wifi/check-saved-profile", handlers.CheckProfileHandler)
	http.HandleFunc("/api/wifi/connect", handlers.ConnectHandler)
	http.HandleFunc("/api/wifi/scan", handlers.ScanHandler)

	http.HandleFunc("/library/test/success.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Success"))
	})
	http.HandleFunc("/hotspot-detect.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Success"))
	})

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on port 8080...")
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
