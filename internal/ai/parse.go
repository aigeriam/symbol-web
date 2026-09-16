package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

// parseLines splits raw LLM output into a clean list of non-empty lines,
// stripping common list decorations ("1.", "-", "*") and truncating to max
// items. This is deliberately forgiving: LLMs rarely follow "one per line"
// formatting instructions perfectly.
func parseLines(raw string, max int) []string {
	var out []string
	bulletPrefix := regexp.MustCompile(`^\s*(\d+[\.\)]|[-*•])\s*`)

	for _, line := range strings.Split(raw, "\n") {
		line = bulletPrefix.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		line = strings.Trim(line, `"`)
		if line == "" {
			continue
		}
		out = append(out, line)
		if len(out) >= max {
			break
		}
	}
	return out
}

// parseVariationsJSON extracts a JSON array of Variation objects from raw
// model output, tolerating surrounding prose or markdown code fences.
func parseVariationsJSON(raw string) ([]Variation, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end < start {
		return nil, ErrBadResponse
	}
	jsonSlice := raw[start : end+1]

	var variations []Variation
	if err := json.Unmarshal([]byte(jsonSlice), &variations); err != nil {
		return nil, err
	}

	// Normalize/validate banner names so the frontend never receives garbage.
	for i := range variations {
		if !isValidBannerName(variations[i].SuggestedBanner) {
			variations[i].SuggestedBanner = BannerStandard
		}
	}
	return variations, nil
}

func isValidBannerName(name string) bool {
	switch name {
	case BannerStandard, BannerShadow, BannerThinkertoy:
		return true
	default:
		return false
	}
}
