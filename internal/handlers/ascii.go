package handlers

import (
	"net/http"
	"text/template"
)

type AsciiHandler struct {
	Generator *ascii.Generator
	Templates *template.Template
}

func NewAsciiHandler(generator *ascii.Generator, tmpl *template.Template) *AsciiHandler {
	return &AsciiHandler{
		generator: generator,
		tmpl:      tmpl,
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
}
