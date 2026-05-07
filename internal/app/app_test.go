package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"

	"ugfetch/internal/cli"
)

func TestRunWritesSanityOutput(t *testing.T) {
	dir := t.TempDir()
	var printed []string

	deps := DefaultDependencies()
	deps.FetchTab = func(id int64) (ultimateguitar.TabResult, error) {
		if id != 98765 {
			t.Fatalf("FetchTab id = %d, want %d", id, 98765)
		}
		return ultimateguitar.TabResult{
			ArtistName: "AC/DC",
			SongName:   "Thunder Struck",
			Content: strings.Join([]string{
				"Credits",
				"[Intro]",
				"C   G | Am  F",
				"Lyrics",
			}, "\n"),
		}, nil
	}
	deps.Println = func(args ...any) (int, error) {
		printed = append(printed, argsToString(args...))
		return 1, nil
	}

	opts := cli.Options{
		ID:        98765,
		Key:       "D",
		OutputDir: dir,
	}
	if err := Run(opts, deps); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	outFile := filepath.Join(dir, "AC-DC", "thunder-struck.md")
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want := "[Intro]\nD   A | Bm  G\nLyrics\n"
	if string(got) != want {
		t.Fatalf("output file = %q, want %q", string(got), want)
	}
	if len(printed) != 1 || printed[0] != outFile {
		t.Fatalf("printed output = %v, want [%q]", printed, outFile)
	}
}

func argsToString(args ...any) string {
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(fmt.Sprint(arg))
	}
	return b.String()
}
