package cli

import "testing"

func TestParseArgs(t *testing.T) {
	t.Run("numeric id and flags", func(t *testing.T) {
		opts, err := ParseArgs([]string{"12345", "--key", "G", "--markup", "--output-dir", "./tabs"})
		if err != nil {
			t.Fatalf("ParseArgs() error = %v", err)
		}
		if opts.ID != 12345 {
			t.Fatalf("ID = %d, want %d", opts.ID, 12345)
		}
		if opts.Key != "G" {
			t.Fatalf("Key = %q, want %q", opts.Key, "G")
		}
		if !opts.Markup {
			t.Fatalf("Markup = false, want true")
		}
		if opts.OutputDir != "./tabs" {
			t.Fatalf("OutputDir = %q, want %q", opts.OutputDir, "./tabs")
		}
	})

	t.Run("url with embedded id", func(t *testing.T) {
		opts, err := ParseArgs([]string{"https://www.ultimate-guitar.com/tab/example/song_98765"})
		if err != nil {
			t.Fatalf("ParseArgs() error = %v", err)
		}
		if opts.ID != 98765 {
			t.Fatalf("ID = %d, want %d", opts.ID, 98765)
		}
	})

	t.Run("errors on missing id", func(t *testing.T) {
		_, err := ParseArgs([]string{"--markup"})
		if err == nil {
			t.Fatal("ParseArgs() error = nil, want error")
		}
	})
}
