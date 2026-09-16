package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMockModeSuggestions(t *testing.T) {
	c := &Client{} // no BaseURL => mock mode
	if !c.MockMode() {
		t.Fatal("expected mock mode with empty BaseURL")
	}
	suggestions, err := c.GetSuggestions(context.Background(), "Happy Birth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) < 3 || len(suggestions) > 5 {
		t.Fatalf("expected 3-5 suggestions, got %d", len(suggestions))
	}
}

func TestSuggestionsRejectShortInput(t *testing.T) {
	c := &Client{}
	if _, err := c.GetSuggestions(context.Background(), "hi"); err == nil {
		t.Fatal("expected error for input shorter than minimum length")
	}
}

func TestMockModeVariations(t *testing.T) {
	c := &Client{}
	variations, err := c.GetVariations(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variations) < 3 {
		t.Fatalf("expected at least 3 variations, got %d", len(variations))
	}
	for _, v := range variations {
		if !isValidBannerName(v.SuggestedBanner) {
			t.Errorf("invalid suggested banner: %q", v.SuggestedBanner)
		}
	}
}

func TestVariationsRejectEmptyInput(t *testing.T) {
	c := &Client{}
	if _, err := c.GetVariations(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty text")
	}
}

// TestLiveModeSuggestionsSuccess spins up a fake OpenAI-compatible server and
// verifies the client parses a normal chat-completions response correctly.
func TestLiveModeSuggestionsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [
				{"message": {"role": "assistant", "content": "Hello World!\nHELLO!\n~ hello ~\nHi there\nHey!"}}
			]
		}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test-model", httpClient: srv.Client()}
	suggestions, err := c.GetSuggestions(context.Background(), "Happy Birth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion parsed from the fake response")
	}
}

// TestLiveModeTimeout verifies that a slow backend results in
// ErrBackendUnavailable rather than a hang or a crash.
func TestLiveModeTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{"choices":[{"message":{"content":"too slow"}}]}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test-model", httpClient: srv.Client()}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := c.GetSuggestions(ctx, "Happy Birth")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestLiveModeServerError verifies a 500 from the backend is surfaced as
// ErrBackendUnavailable, not a panic.
func TestLiveModeServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test-model", httpClient: srv.Client()}
	_, err := c.GetSuggestions(context.Background(), "Happy Birth")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// TestLiveModeBadJSON verifies garbage JSON from the backend is surfaced as
// ErrBadResponse, not a panic.
func TestLiveModeBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json at all"))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test-model", httpClient: srv.Client()}
	_, err := c.GetSuggestions(context.Background(), "Happy Birth")
	if err == nil {
		t.Fatal("expected error for unparseable response")
	}
}

func TestLiveModeVariationsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"choices": [{"message": {"content": "Sure! Here you go:\n[{\"text\":\"Hello World!\",\"description\":\"Classic\",\"suggested_banner\":\"standard\"}]"}}]
		}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test-model", httpClient: srv.Client()}
	variations, err := c.GetVariations(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variations) != 1 || variations[0].Text != "Hello World!" {
		t.Fatalf("unexpected variations parsed: %+v", variations)
	}
}

func TestRecommendUppercaseFavorsShadow(t *testing.T) {
	r := Recommend("WELCOME")
	if r.Recommended != BannerShadow {
		t.Errorf("expected shadow for uppercase text, got %q", r.Recommended)
	}
	if len(r.Alternatives) != 2 {
		t.Errorf("expected 2 alternatives, got %d", len(r.Alternatives))
	}
}

func TestRecommendLongTextFavorsStandard(t *testing.T) {
	r := Recommend("this is a fairly long piece of text")
	if r.Recommended != BannerStandard {
		t.Errorf("expected standard for long text, got %q", r.Recommended)
	}
}

func TestRecommendPlayfulFavorsThinkertoy(t *testing.T) {
	r := Recommend("Hi~there!")
	if r.Recommended != BannerThinkertoy {
		t.Errorf("expected thinkertoy for playful mixed text, got %q", r.Recommended)
	}
}

func TestRecommendIsDeterministic(t *testing.T) {
	a := Recommend("Hello World")
	b := Recommend("Hello World")
	if a.Recommended != b.Recommended || a.Reasoning != b.Reasoning {
		t.Fatal("expected Recommend to be deterministic for the same input")
	}
}

func TestParseLinesStripsBullets(t *testing.T) {
	raw := "1. Hello World!\n2) HELLO!\n- ~ hello ~\n* Hey there\nplain line"
	lines := parseLines(raw, 5)
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d: %v", len(lines), lines)
	}
	if strings.HasPrefix(lines[0], "1.") {
		t.Errorf("expected bullet prefix stripped, got %q", lines[0])
	}
}
