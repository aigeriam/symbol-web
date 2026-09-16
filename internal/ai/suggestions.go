package ai

import (
	"context"
	"fmt"
	"strings"
)

const MinSuggestInputLength = 3

const (
	BannerStandard   = "standard"
	BannerShadow     = "shadow"
	BannerThinkertoy = "thinkertoy"
)

type Variation struct {
	Text            string `json:"text"`
	Description     string `json:"description"`
	SuggestedBanner string `json:"suggested_banner"`
}

func mockSuggestions(text string) []string {
	base := strings.TrimSpace(text)
	return []string{
		base + "!",
		base + " Team!",
		base + " [Name]",
		"~ " + base + " ~",
		base + " 2026",
	}
}
func (c *Client) GetSuggestions(ctx context.Context, text string) ([]string, error) {
	if len(strings.TrimSpace(text)) < MinSuggestInputLength {
		return nil, fmt.Errorf("text must be at least %d characters", MinSuggestInputLength)
	}

	if c.MockMode() {
		return mockSuggestions(text), nil
	}
	prompt := fmt.Sprintf(
		"Complete this text creatively for ASCII art display:\n"+
			"Input: %q\n"+
			"Provide 3-5 short, creative completions suitable for ASCII art.\n"+
			"Each completion should be under 50 characters.\n"+
			"Return only the completions, one per line.",
		text,
	)
	raw, err := c.complete(ctx, prompt)
	if err != nil {
		return nil, err
	}
	suggestions := parseLines(raw, 5)
	if len(suggestions) == 0 {
		return nil, fmt.Errorf("%w: model returned no usable suggestions", ErrBadResponse)
	}
	return suggestions, nil
}
func (c *Client) GetVariations(ctx context.Context, text string) ([]Variation, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text must not be empty")
	}

	if c.MockMode() {
		return mockVariations(text), nil
	}

	prompt := fmt.Sprintf(
		"Generate creative variations of this text for ASCII art:\n"+
			"Input: %q\n\n"+
			"Create 4 variations:\n"+
			"1. Professional/formal style\n"+
			"2. Bold/emphatic style\n"+
			"3. Friendly/casual style\n"+
			"4. Artistic/decorative style\n\n"+
			"For each variation, provide:\n"+
			"- The modified text\n"+
			"- Brief description (under 30 chars)\n"+
			"- Recommended banner (shadow/standard/thinkertoy)\n\n"+
			"Respond ONLY with a JSON array of objects shaped like:\n"+
			`[{"text": "...", "description": "...", "suggested_banner": "standard"}]`,
		text,
	)

	raw, err := c.complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	variations, err := parseVariationsJSON(raw)
	if err != nil || len(variations) == 0 {
		// The model didn't give us clean JSON. Rather than fail the request,
		// fall back to a deterministic transformation of its raw text so the
		// feature stays usable; but surface a bad-response error if we truly
		// got nothing.
		return nil, fmt.Errorf("%w: could not parse model output as variations", ErrBadResponse)
	}
	return variations, nil
}

// mockSuggestions produces deterministic, testable output without any

// mockVariations produces deterministic, testable output without any
// network access.
func mockVariations(text string) []Variation {
	base := strings.TrimSpace(text)
	title := base
	if len(base) > 0 {
		title = strings.ToUpper(string(base[0])) + strings.ToLower(base[1:])
	}
	return []Variation{
		{
			Text:            title,
			Description:     "Professional style",
			SuggestedBanner: BannerStandard,
		},
		{
			Text:            strings.ToUpper(base) + "!",
			Description:     "Bold, attention-grabbing",
			SuggestedBanner: BannerShadow,
		},
		{
			Text:            "hey " + strings.ToLower(base) + " :)",
			Description:     "Friendly, casual style",
			SuggestedBanner: BannerStandard,
		},
		{
			Text:            "~ " + base + " ~",
			Description:     "Artistic, decorated style",
			SuggestedBanner: BannerThinkertoy,
		},
	}
}
