package cli

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const UsageText = "usage: ugfetch <ug-id|song-url> [--key KEY] [--markup] [--output-dir PATH]"

type Options struct {
	ID        int64
	Key       string
	Markup    bool
	OutputDir string
}

var trailingIDRe = regexp.MustCompile(`(\d+)$`)

func ParseArgs(args []string) (Options, error) {
	if len(args) == 0 {
		return Options{}, errors.New(UsageText)
	}

	var opts Options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return Options{}, errors.New(UsageText)
		case arg == "--key":
			if i+1 >= len(args) {
				return Options{}, errors.New("--key requires a value")
			}
			opts.Key = args[i+1]
			i++
		case strings.HasPrefix(arg, "--key="):
			opts.Key = strings.TrimPrefix(arg, "--key=")
		case arg == "--markup":
			opts.Markup = true
		case arg == "--output-dir":
			if i+1 >= len(args) {
				return Options{}, errors.New("--output-dir requires a value")
			}
			opts.OutputDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output-dir="):
			opts.OutputDir = strings.TrimPrefix(arg, "--output-dir=")
		case strings.HasPrefix(arg, "-"):
			return Options{}, fmt.Errorf("unknown flag: %s", arg)
		default:
			if opts.ID != 0 {
				return Options{}, fmt.Errorf("unexpected extra argument: %s", arg)
			}
			parsed, err := parseUGTarget(arg)
			if err != nil {
				return Options{}, err
			}
			opts.ID = parsed
		}
	}

	if opts.ID == 0 {
		return Options{}, errors.New("missing ug id")
	}

	return opts, nil
}

func parseUGTarget(arg string) (int64, error) {
	if parsed, err := strconv.ParseInt(arg, 10, 64); err == nil {
		return parsed, nil
	}

	if !strings.Contains(arg, "://") {
		arg = "https://" + strings.TrimPrefix(arg, "//")
	}

	parsedURL, err := url.Parse(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid ug id or URL %q: %w", arg, err)
	}

	candidates := []string{parsedURL.Path, parsedURL.Query().Get("tab")}
	for _, candidate := range candidates {
		if id, ok := extractTrailingID(candidate); ok {
			return id, nil
		}
	}

	return 0, fmt.Errorf("could not extract UG id from URL %q", arg)
}

func extractTrailingID(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	matches := trailingIDRe.FindStringSubmatch(value)
	if matches == nil {
		return 0, false
	}
	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
