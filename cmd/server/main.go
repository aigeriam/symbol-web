package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"symbol-web/internal/ai"
	"symbol-web/internal/ascii"
	"symbol-web/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	generator := ascii.Newgenerator(".")
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
	aiClient := ai.NewClientFromEnv()
	if aiClient.MockMode() {
		log.Println("LLM_BASE_URL not set — running AI endpoints in mock mode (no API key needed).")

	} else {
		log.Printf("LLM backend configured: %s (model: %s)", aiClient.BaseURL, aiClient.Model)
	}
	asciiHandler := handlers.NewAsciiHandler(generator, tmpl)
	aiHandler := handlers.NewAiHandler(aiClient)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.HandleFunc("/", asciiHandler.Index)
	mux.HandleFunc("/symbol-art", asciiHandler.GenerateArt)
	mux.HandleFunc("/api/suggest/", aiHandler.Suggest)
	mux.HandleFunc("/api/recommend-banner", aiHandler.RecommendBanner)
	mux.HandleFunc("/api/variations", aiHandler.Variations)
}