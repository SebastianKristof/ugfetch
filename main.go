package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ugfetch/internal/cli"
	"ugfetch/internal/tab"
)

func main() {
	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fatal(err)
	}

	if opts.OutputDir != "" {
		opts.OutputDir, err = filepath.Abs(opts.OutputDir)
		if err != nil {
			fatal(err)
		}
	}

	tabResult, err := tab.FetchTab(opts.ID)
	if err != nil {
		fatal(err)
	}

	content := tab.StripLeadingMetadata(strings.TrimRight(tabResult.Content, "\n"))
	if !opts.Markup {
		content = tab.StripUGMarkup(content)
	}

	if opts.Key != "" {
		sourceKey := tab.TabSourceKey(tabResult, content)
		if opts.Markup {
			content, err = tab.TransposeMarkup(content, sourceKey, opts.Key)
		} else {
			content, err = tab.Transpose(content, sourceKey, opts.Key)
		}
		if err != nil {
			fatal(err)
		}
	}

	baseDir := opts.OutputDir
	if baseDir == "" {
		baseDir, err = os.Getwd()
		if err != nil {
			fatal(err)
		}
	}

	artistDir := filepath.Join(baseDir, tab.SafeComponent(tabResult.ArtistName))
	if err := os.MkdirAll(artistDir, 0o755); err != nil {
		fatal(err)
	}

	outFile := filepath.Join(artistDir, tab.Slugify(tabResult.SongName)+".md")
	if err := os.WriteFile(outFile, []byte(content+"\n"), 0o644); err != nil {
		fatal(err)
	}

	fmt.Println(outFile)
}

func fatal(err error) {
	if err == nil {
		return
	}
	if err.Error() == cli.UsageText {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
