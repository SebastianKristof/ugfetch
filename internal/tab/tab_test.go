package tab

import (
	"strings"
	"testing"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"
)

func TestSlugify(t *testing.T) {
	got := Slugify("Crème Brûlée / Live!")
	want := "creme-brulee-live"
	if got != want {
		t.Fatalf("Slugify() = %q, want %q", got, want)
	}
}

func TestSafeComponent(t *testing.T) {
	got := SafeComponent("AC/DC: Live")
	want := "AC-DC- Live"
	if got != want {
		t.Fatalf("SafeComponent() = %q, want %q", got, want)
	}
}

func TestStripLeadingMetadata(t *testing.T) {
	got := StripLeadingMetadata("Title\nArtist\n\n[Intro]\nC G Am F")
	want := "[Intro]\nC G Am F"
	if got != want {
		t.Fatalf("StripLeadingMetadata() = %q, want %q", got, want)
	}
}

func TestTranspose(t *testing.T) {
	got, err := Transpose("[Intro]\nC   G | Am  F\nLyrics", "C", "D")
	if err != nil {
		t.Fatalf("Transpose() error = %v", err)
	}
	want := "[Intro]\nD   A | Bm  G\nLyrics"
	if got != want {
		t.Fatalf("Transpose() = %q, want %q", got, want)
	}
}

func TestTransposeMarkup(t *testing.T) {
	got, err := TransposeMarkup("[ch]C[/ch] [ch]G/B[/ch]", "C", "D")
	if err != nil {
		t.Fatalf("TransposeMarkup() error = %v", err)
	}
	want := "[ch]D[/ch] [ch]A/C#[/ch]"
	if got != want {
		t.Fatalf("TransposeMarkup() = %q, want %q", got, want)
	}
}

func TestTabSourceKey(t *testing.T) {
	tabResult := ultimateguitar.TabResult{
		TonalityName: "",
		Content:      "",
	}

	got := TabSourceKey(tabResult, strings.Join([]string{
		"[Intro]",
		"C G Am F",
	}, "\n"))
	want := "C"
	if got != want {
		t.Fatalf("TabSourceKey() = %q, want %q", got, want)
	}
}
