package ascii

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	BannerShadow     = "shadow"
	BannerStandard   = "standard"
	BannerThinkertoy = "thinkertoy"
)

type Generator struct {
	mu    sync.RWMutex
	dir   string
	fonts map[string]*font
}
type font struct {
	height int
	glyphs map[rune][]string
}

var ValidBanners = []string{BannerStandard, BannerShadow, BannerThinkertoy}
var ErrInvalidBanner = errors.New("ascii: invalid banner name")
var ErrEmptyText = errors.New("ascii: text must not be empty")

// to say we are at the current working directory
func Newgenerator(dir string) *Generator {
	return &Generator{
		dir:   dir,
		fonts: make(map[string]*font),
	}
}
func IsValidBanner(name string) bool {
	for _, b := range ValidBanners {
		if b == name {
			return true
		}
	}
	return false
}
func (g *Generator) Render(text, banner string) (string, error) {
	if text == "" {
		return "", ErrEmptyText
	}
	if !IsValidBanner(banner) {
		return "", fmt.Errorf("%w: %q", ErrInvalidBanner, banner)
	}

	font, err := g.loadFont(banner)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			output.WriteByte('\n')
			continue
		}

		rendered, err := renderLine(font, line)
		if err != nil {
			return "", err
		}
		output.WriteString(rendered)
	}
	return output.String(), nil
}
func renderLine(f *font, line string) (string, error) {
	rows := make([]string, f.height)
	for _, r := range line {
		glyph := f.glyphs[r]
		if glyph == nil {
			glyph = f.glyphs[' ']
		}
		for i := 0; i < f.height; i++ {
			rows[i] += glyph[i]
		}
	}

	var output strings.Builder
	for _, row := range rows {
		output.WriteString(row)
		output.WriteByte('\n')
	}
	return output.String(), nil
}
func (g *Generator) loadFont(banner string) (*font, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if font, exists := g.fonts[banner]; exists {
		return font, nil
	}

	path := filepath.Join(g.dir, banner+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ascii: reading banner file %q: %w", path, err)
	}

	font, err := parseFont(string(data))
	if err != nil {
		return nil, fmt.Errorf("ascii: parsing banner file %q: %w", path, err)
	}

	g.fonts[banner] = font
	return font, nil
}
func parseFont(data string) (*font, error) {
	data = strings.TrimRight(data, "\n")
	blocks := strings.Split(data, "\n\n")
	if len(blocks) != 95 {
		return nil, fmt.Errorf("expected 95 character blocks (codes 32-126), got %d", len(blocks))
	}

	height := 0
	for _, b := range blocks {
		h := len(strings.Split(b, "\n"))
		if h > height {
			height = h
		}
	}

	glyphs := make(map[rune][]string, 95)
	for i, b := range blocks {
		code := 32 + i
		lines := strings.Split(b, "\n")
		for len(lines) < height {
			lines = append(lines, "")
		}
		glyphs[rune(code)] = lines
	}

	return &font{height: height, glyphs: glyphs}, nil
}
