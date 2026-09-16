package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"symbol-web/internal/ascii"
	"text/template"
)

type AsciiHandler struct {
	Generator *ascii.Generator
	Templates *template.Template
}
type PageData struct {
	Banners        []string
	SelectedBanner string
	Text           string
	Result         string
	ErrorMessage   string
}

func NewAsciiHandler(generator *ascii.Generator, tmpl *template.Template) *AsciiHandler {
	return &AsciiHandler{
		Generator: generator,
		Templates: tmpl,
	}
}

// main page, sends default banner and passes it to html template
func (h *AsciiHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.renderError(w, http.StatusNotFound, "Page not found")
		return
	}
	if r.Method != http.MethodGet {
		h.renderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	data := PageData{
		Banners:        ascii.ValidBanners,
		SelectedBanner: ascii.BannerStandard,
	}
	h.render(w, http.StatusOK, "index.html", data)
}

// post request, reads banner choice and text
func (h *AsciiHandler) GenerateArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.renderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, http.StatusBadRequest, "Could not parse form data")
		return
	}

	text := r.FormValue("text")
	banner := r.FormValue("banner")

	data := PageData{
		Banners:        ascii.ValidBanners,
		SelectedBanner: banner,
		Text:           text,
	}

	if strings.TrimSpace(text) == "" {
		data.ErrorMessage = "Text input must not be empty."
		h.render(w, http.StatusBadRequest, "index.html", data)
		return
	}
	if len(text) > 1000 {
		data.ErrorMessage = "Text input is too long (maximum 1000 characters)."
		h.render(w, http.StatusBadRequest, "index.html", data)
		return
	}
	if banner == "" {
		banner = ascii.BannerStandard
		data.SelectedBanner = banner
	}
	if !ascii.IsValidBanner(banner) {
		data.ErrorMessage = "Invalid banner selection: " + banner
		h.render(w, http.StatusBadRequest, "index.html", data)
		return
	}

	result, err := h.Generator.Render(text, banner)
	if err != nil {
		log.Printf("ascii render error: %v", err)
		if errors.Is(err, ascii.ErrEmptyText) || errors.Is(err, ascii.ErrInvalidBanner) {
			data.ErrorMessage = err.Error()
			h.render(w, http.StatusBadRequest, "index.html", data)
			return
		}
		data.ErrorMessage = "Something went wrong generating your ASCII art. Please try again."
		h.render(w, http.StatusInternalServerError, "index.html", data)
		return
	}

	data.Result = result
	h.render(w, http.StatusOK, "index.html", data)
}
func (h *AsciiHandler) render(w http.ResponseWriter, status int, name string, data any) {
	if h.Templates.Lookup(name) == nil {
		h.renderError(w, http.StatusNotFound, "Template not found: "+name)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.Templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template execution error for %q: %v", name, err)
		// Headers are already sent at this point in many cases; best effort.
		http.Error(w, "Internal server error rendering page", http.StatusInternalServerError)
	}
}
func (h *AsciiHandler) renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if h.Templates.Lookup("error.html") == nil {
		http.Error(w, message, status)
		return
	}
	_ = h.Templates.ExecuteTemplate(w, "error.html", map[string]any{
		"StatusCode": status,
		"Message":    message,
	})
}
