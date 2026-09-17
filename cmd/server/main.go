package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"symbol-web/internal/ai"
	"symbol-web/internal/ascii"
	"symbol-web/internal/handlers"
	"text/template"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running at http://localhost:%s", port)

	root := projectRoot()
	generator := ascii.Newgenerator(root)

	tmplPath := filepath.Join(root, "templates", "*.html")
	tmpl, err := template.ParseGlob(tmplPath)
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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(root, "static")))))
	mux.HandleFunc("/", asciiHandler.Index)
	mux.HandleFunc("/symbol-art", asciiHandler.GenerateArt)
	mux.HandleFunc("/api/suggest/", aiHandler.Suggest)
	mux.HandleFunc("/api/recommend-banner", aiHandler.RecommendBanner)
	mux.HandleFunc("/api/variations", aiHandler.Variations)

	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
