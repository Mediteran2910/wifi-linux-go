package main

import (
	"log"
	"net/http"
	"time"

	"my-go-app/handlers"
	softap "my-go-app/soft-ap"
)

func main() {
	iface := "wlxec750c9f9f9b"

	if err := softap.StartHotspot(iface); err != nil {
		log.Fatalf("Fatal: Failed to start the hotspot: %v", err)
	}

	// Captive portal endpoints (Android, iOS, Windows)
	http.HandleFunc("/generate_204", captivePortalHandler)
	http.HandleFunc("/gen_204", captivePortalHandler)
	http.HandleFunc("/hotspot-detect.html", captivePortalHandler)
	http.HandleFunc("/library/test/success.html", captivePortalHandler)
	http.HandleFunc("/ncsi.txt", captivePortalHandler)
	http.HandleFunc("/connecttest.txt", captivePortalHandler)
	http.HandleFunc("/redirect", captivePortalHandler)

	// Static portal files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// API endpoints
	http.HandleFunc("/api/wifi/check-saved-profile", handlers.CheckProfileHandler)
	http.HandleFunc("/api/wifi/connect", handlers.ConnectHandler)
	http.HandleFunc("/api/wifi/scan", handlers.ScanHandler)

	// HTTP server
	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on port 8080...")
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// captivePortalHandler redirects devices to the portal
func captivePortalHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Captive portal probe detected: %s from %s", r.URL.Path, r.RemoteAddr)
	http.Redirect(w, r, "http://10.42.0.1:8080/", http.StatusFound)
}
