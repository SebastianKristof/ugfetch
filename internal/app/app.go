package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"

	"ugfetch/internal/cli"
	"ugfetch/internal/tab"
)

type Dependencies struct {
	FetchTab  func(int64) (ultimateguitar.TabResult, error)
	Getwd     func() (string, error)
	Abs       func(string) (string, error)
	MkdirAll  func(string, os.FileMode) error
	WriteFile func(string, []byte, os.FileMode) error
	Println   func(...any) (int, error)
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		FetchTab:  tab.FetchTab,
		Getwd:     os.Getwd,
		Abs:       filepath.Abs,
		MkdirAll:  os.MkdirAll,
		WriteFile: os.WriteFile,
		Println:   fmt.Println,
	}
}

func Run(opts cli.Options, deps Dependencies) error {
	if deps.FetchTab == nil {
		deps.FetchTab = tab.FetchTab
	}
	if deps.Getwd == nil {
		deps.Getwd = os.Getwd
	}
	if deps.Abs == nil {
		deps.Abs = filepath.Abs
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.WriteFile == nil {
		deps.WriteFile = os.WriteFile
	}
	if deps.Println == nil {
		deps.Println = fmt.Println
	}

	if opts.OutputDir != "" {
		absDir, err := deps.Abs(opts.OutputDir)
		if err != nil {
			return err
		}
		opts.OutputDir = absDir
	}

	tabResult, err := deps.FetchTab(opts.ID)
	if err != nil {
		return err
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
			return err
		}
	}

	baseDir := opts.OutputDir
	if baseDir == "" {
		baseDir, err = deps.Getwd()
		if err != nil {
			return err
		}
	}

	artistDir := filepath.Join(baseDir, tab.SafeComponent(tabResult.ArtistName))
	if err := deps.MkdirAll(artistDir, 0o755); err != nil {
		return err
	}

	outFile := filepath.Join(artistDir, tab.Slugify(tabResult.SongName)+".md")
	if err := deps.WriteFile(outFile, []byte(content+"\n"), 0o644); err != nil {
		return err
	}

	if _, err := deps.Println(outFile); err != nil {
		return err
	}

	return nil
}
