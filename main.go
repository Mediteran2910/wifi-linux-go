package main

import (
	"log"
	"net/http"
	"time"

	"my-go-app/handlers"
)

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.HandleFunc("/api/wifi/check-saved-profile", handlers.CheckProfileHandler)
	http.HandleFunc("/api/wifi/connect", handlers.ConnectHandler)
	http.HandleFunc("/api/wifi/scan", handlers.ScanHandler)

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
