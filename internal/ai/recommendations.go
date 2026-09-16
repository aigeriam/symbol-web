package ai

import (
	"sort"
	"strings"
	"unicode"
)

// Alternative is one non-recommended banner option, with a score in [0,1]
// and a short human-readable reason.
type Alternative struct {
	Banner string  `json:"banner"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// Recommendation is the full response of the rule-based recommender.
type Recommendation struct {
	Recommended  string        `json:"recommended"`
	Reasoning    string        `json:"reasoning"`
	Alternatives []Alternative `json:"alternatives"`
}

// textProfile captures the handful of text characteristics the rules below
// look at. This reuses the "profiling" idea from symbol-fs: cheap, explainable
// signals rather than anything probabilistic.
type textProfile struct {
	length          int
	isAllUpper      bool
	isAllLower      bool
	hasSpecialChars bool
	wordCount       int
}

func profile(text string) textProfile {
	trimmed := strings.TrimSpace(text)
	p := textProfile{
		length:    len([]rune(trimmed)),
		wordCount: len(strings.Fields(trimmed)),
	}

	hasLetter := false
	allUpper := true
	allLower := true
	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsUpper(r) {
				allLower = false
			} else if unicode.IsLower(r) {
				allUpper = false
			}
		} else if !unicode.IsSpace(r) && !unicode.IsDigit(r) {
			p.hasSpecialChars = true
		}
	}
	p.isAllUpper = hasLetter && allUpper
	p.isAllLower = hasLetter && allLower

	return p
}

// scoreBanner returns a 0-1 fit score for each banner given the text's
// profile. These are the same deterministic, explainable rules described in
// the assignment, expressed as scores so we can rank alternatives instead of
// only picking a single winner.
func scoreBanner(p textProfile) map[string]float64 {
	scores := map[string]float64{
		BannerStandard:   0.5,
		BannerShadow:     0.5,
		BannerThinkertoy: 0.5,
	}

	// Rule 1: all-uppercase or very short text favors bold "shadow".
	if p.isAllUpper || p.length <= 6 {
		scores[BannerShadow] += 0.4
	} else {
		scores[BannerShadow] -= 0.15
	}

	// Rule 2: long text favors the more readable "standard".
	if p.length > 12 {
		scores[BannerStandard] += 0.4
	} else {
		scores[BannerStandard] -= 0.1
	}

	// Rule 3: playful or symbol-heavy short/mixed-case text favors "thinkertoy".
	if p.hasSpecialChars {
		scores[BannerThinkertoy] += 0.3
	}
	if !p.isAllUpper && !p.isAllLower && p.length <= 12 {
		scores[BannerThinkertoy] += 0.2
	}
	if p.isAllUpper {
		scores[BannerThinkertoy] -= 0.2
	}

	// Clamp to [0, 1].
	for k, v := range scores {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		scores[k] = v
	}
	return scores
}

// reasonFor returns a short, human-readable explanation for why a given
// banner scored the way it did against this text.
func reasonFor(banner string, p textProfile, isTop bool) string {
	switch banner {
	case BannerShadow:
		if p.isAllUpper {
			return "Bold, uppercase text works best with shadow for maximum impact and readability."
		}
		if p.length <= 6 {
			return "Short text stands out with shadow's bold look."
		}
		if isTop {
			return "Shadow gives this text strong visual weight."
		}
		return "Less ideal here since the text isn't short or uppercase."
	case BannerStandard:
		if p.length > 12 {
			return "Readable for longer text, keeps ASCII art from getting too wide per line."
		}
		if isTop {
			return "A safe, highly readable default for this text."
		}
		return "Good alternative, slightly less impactful than the top pick."
	case BannerThinkertoy:
		if p.hasSpecialChars {
			return "Symbol-heavy text suits thinkertoy's playful, decorative style."
		}
		if p.isAllUpper {
			return "Too decorative for uppercase text; shadow reads better."
		}
		if isTop {
			return "Playful style fits this short, mixed-case text."
		}
		return "A more decorative option if you want a playful feel."
	}
	return ""
}

// Recommend analyzes text using rules only (no LLM, no network) and returns
// the best banner plus ranked alternatives.
func Recommend(text string) Recommendation {
	p := profile(text)
	scores := scoreBanner(p)

	type kv struct {
		Banner string
		Score  float64
	}
	ranked := make([]kv, 0, len(scores))
	for b, s := range scores {
		ranked = append(ranked, kv{b, s})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		// Stable, deterministic tie-break by banner name.
		return ranked[i].Banner < ranked[j].Banner
	})

	top := ranked[0]
	alts := make([]Alternative, 0, len(ranked)-1)
	for _, kv := range ranked[1:] {
		alts = append(alts, Alternative{
			Banner: kv.Banner,
			Score:  round2(kv.Score),
			Reason: reasonFor(kv.Banner, p, false),
		})
	}

	return Recommendation{
		Recommended:  top.Banner,
		Reasoning:    reasonFor(top.Banner, p, true),
		Alternatives: alts,
	}
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
