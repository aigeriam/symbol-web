package main

import (
	"html/template"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	generator := ascii.NewGenerator(".")
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
	aiClient := ai.NewClientFromEnv()
	if aiClient.MockMode(){
		log.Println("LLM_BASE_URL not set — running AI endpoints in mock mode (no API key needed).")

	}else{
		log.Printf("LLM backend configured: %s (model: %s)", aiClient.BaseURL, aiClient.Model)
	}
	asciiHandler:=handlers.NewAsciiHandler(generator, tmpl)
	aiHandler:=handlers.NewAIHandler(aiClient) 
	mux:=http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.Handle("/ascii/", asciiHandler)
	mux.Handle("/ai/", aiHandler)
	

}
