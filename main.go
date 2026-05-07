package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"
)

var chordRe = regexp.MustCompile(`^([A-G](?:#|b)?)(m7b5|mmaj7|maj13|maj9|maj7|m13|m11|m9|m7|m6|7sus4|7sus2|sus4|sus2|add11|add9|dim7|dim|aug|maj|m|6|7|9|11|13|5)?(?:/([A-G](?:#|b)?))?$`)
var splitRe = regexp.MustCompile(`(\s+)`)

var semitones = map[string]int{
	"C": 0, "B#": 0,
	"C#": 1, "Db": 1,
	"D":  2,
	"D#": 3, "Eb": 3,
	"E": 4, "Fb": 4,
	"E#": 5, "F": 5,
	"F#": 6, "Gb": 6,
	"G":  7,
	"G#": 8, "Ab": 8,
	"A":  9,
	"A#": 10, "Bb": 10,
	"B": 11, "Cb": 11,
}

var sharpNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var flatNames = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
var markupReplacer = strings.NewReplacer("[tab]", "", "[/tab]", "", "[ch]", "", "[/ch]", "")
var markupChordRe = regexp.MustCompile(`\[ch\](.*?)\[/ch\]`)

const usageText = "usage: ugfetch <ug-id|song-url> [--key KEY] [--markup] [--output-dir PATH]"

func main() {
	id, targetKey, showMarkup, outputDir, err := parseArgs(os.Args[1:])
	if err != nil {
		fatal(err)
	}

	if outputDir != "" {
		outputDir, err = filepath.Abs(outputDir)
		if err != nil {
			fatal(err)
		}
	}

	tab, err := fetchTab(id)
	if err != nil {
		fatal(err)
	}

	content := stripLeadingMetadata(strings.TrimRight(tab.Content, "\n"))
	if !showMarkup {
		content = stripUGMarkup(content)
	}

	if targetKey != "" {
		sourceKey := tabSourceKey(tab, content)
		if showMarkup {
			content, err = transposeMarkup(content, sourceKey, targetKey)
		} else {
			content, err = transpose(content, sourceKey, targetKey)
		}
		if err != nil {
			fatal(err)
		}
	}

	baseDir := outputDir
	if baseDir == "" {
		baseDir, err = os.Getwd()
		if err != nil {
			fatal(err)
		}
	}

	artistDir := filepath.Join(baseDir, safeComponent(tab.ArtistName))
	if err := os.MkdirAll(artistDir, 0o755); err != nil {
		fatal(err)
	}

	outFile := filepath.Join(artistDir, slugify(tab.SongName)+".md")
	if err := os.WriteFile(outFile, []byte(content+"\n"), 0o644); err != nil {
		fatal(err)
	}

	fmt.Println(outFile)
}

func parseArgs(args []string) (int64, string, bool, string, error) {
	if len(args) == 0 {
		return 0, "", false, "", errors.New(usageText)
	}

	var id int64
	var key string
	var showMarkup bool
	var outputDir string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return 0, "", false, "", errors.New(usageText)
		case arg == "--key":
			if i+1 >= len(args) {
				return 0, "", false, "", errors.New("--key requires a value")
			}
			key = args[i+1]
			i++
		case strings.HasPrefix(arg, "--key="):
			key = strings.TrimPrefix(arg, "--key=")
		case arg == "--markup":
			showMarkup = true
		case arg == "--output-dir":
			if i+1 >= len(args) {
				return 0, "", false, "", errors.New("--output-dir requires a value")
			}
			outputDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output-dir="):
			outputDir = strings.TrimPrefix(arg, "--output-dir=")
		case strings.HasPrefix(arg, "-"):
			return 0, "", false, "", fmt.Errorf("unknown flag: %s", arg)
		default:
			if id != 0 {
				return 0, "", false, "", fmt.Errorf("unexpected extra argument: %s", arg)
			}
			parsed, err := parseUGTarget(arg)
			if err != nil {
				return 0, "", false, "", err
			}
			id = parsed
		}
	}

	if id == 0 {
		return 0, "", false, "", errors.New("missing ug id")
	}

	return id, key, showMarkup, outputDir, nil
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
	matches := regexp.MustCompile(`(\d+)$`).FindStringSubmatch(value)
	if matches == nil {
		return 0, false
	}
	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func fetchTab(id int64) (ultimateguitar.TabResult, error) {
	s := ultimateguitar.New()
	return s.GetTabByID(id)
}

func safeComponent(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("/", "-", "\\", "-", ":", "-", "\x00", "").Replace(value)
	if value == "" {
		return "unknown"
	}
	return value
}

func stripLeadingMetadata(content string) string {
	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[start]), "[") {
		start++
	}
	return strings.Join(lines[start:], "\n")
}

func stripUGMarkup(content string) string {
	return markupReplacer.Replace(content)
}

func tabSourceKey(tab ultimateguitar.TabResult, content string) string {
	sourceKey := inferKey(stripUGMarkup(content))
	if sourceKey == "" {
		sourceKey = strings.TrimSpace(tab.TonalityName)
	}
	if sourceKey == "" {
		sourceKey = strings.TrimSpace(tab.Recording.TonalityName)
	}
	return sourceKey
}

func slugify(value string) string {
	replacer := strings.NewReplacer(
		"À", "a", "Á", "a", "Â", "a", "Ã", "a", "Ä", "a", "Å", "a",
		"Æ", "ae",
		"Ç", "c",
		"È", "e", "É", "e", "Ê", "e", "Ë", "e",
		"Ì", "i", "Í", "i", "Î", "i", "Ï", "i",
		"Ñ", "n",
		"Ò", "o", "Ó", "o", "Ô", "o", "Õ", "o", "Ö", "o", "Ø", "o",
		"Œ", "oe",
		"Ù", "u", "Ú", "u", "Û", "u", "Ü", "u",
		"Ý", "y",
		"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
		"æ", "ae",
		"ç", "c",
		"è", "e", "é", "e", "ê", "e", "ë", "e",
		"ì", "i", "í", "i", "î", "i", "ï", "i",
		"ñ", "n",
		"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o", "ø", "o",
		"œ", "oe",
		"ù", "u", "ú", "u", "û", "u", "ü", "u",
		"ý", "y", "ÿ", "y",
	)
	value = replacer.Replace(value)
	value = strings.ToLower(value)
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "song"
	}
	return out
}

func inferKey(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if !isChordLine(line) {
			continue
		}
		for _, token := range strings.Fields(line) {
			if isChordToken(token) {
				if root := chordRoot(token); root != "" {
					return root
				}
			}
		}
	}
	return ""
}

func transpose(content, sourceKey, targetKey string) (string, error) {
	sourceRoot := normalizeRoot(sourceKey)
	targetRoot := normalizeRoot(targetKey)
	if sourceRoot == "" {
		return "", fmt.Errorf("could not determine source key from metadata or content: %q", sourceKey)
	}
	if targetRoot == "" {
		return "", fmt.Errorf("could not determine target key from --key: %q", targetKey)
	}

	sourceSemitone, ok := semitones[sourceRoot]
	if !ok {
		return "", fmt.Errorf("unsupported source key: %s", sourceKey)
	}
	targetSemitone, ok := semitones[targetRoot]
	if !ok {
		return "", fmt.Errorf("unsupported target key: %s", targetKey)
	}

	delta := (targetSemitone - sourceSemitone + 12) % 12
	useFlats := strings.Contains(targetRoot, "b") || isFlatKey(targetRoot)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if isChordLine(line) {
			lines[i] = transposeLine(line, delta, useFlats)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func transposeMarkup(content, sourceKey, targetKey string) (string, error) {
	sourceRoot := normalizeRoot(sourceKey)
	targetRoot := normalizeRoot(targetKey)
	if sourceRoot == "" {
		return "", fmt.Errorf("could not determine source key from metadata or content: %q", sourceKey)
	}
	if targetRoot == "" {
		return "", fmt.Errorf("could not determine target key from --key: %q", targetKey)
	}

	sourceSemitone, ok := semitones[sourceRoot]
	if !ok {
		return "", fmt.Errorf("unsupported source key: %s", sourceKey)
	}
	targetSemitone, ok := semitones[targetRoot]
	if !ok {
		return "", fmt.Errorf("unsupported target key: %s", targetKey)
	}

	delta := (targetSemitone - sourceSemitone + 12) % 12
	useFlats := strings.Contains(targetRoot, "b") || isFlatKey(targetRoot)

	return markupChordRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := markupChordRe.FindStringSubmatch(match)
		if matches == nil {
			return match
		}
		return "[ch]" + transposeToken(matches[1], delta, useFlats) + "[/ch]"
	}), nil
}

func normalizeRoot(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if len(key) >= 2 {
		root := key[:2]
		if _, ok := semitones[root]; ok {
			return root
		}
	}
	if _, ok := semitones[key[:1]]; ok {
		return key[:1]
	}
	return ""
}

func isFlatKey(root string) bool {
	switch root {
	case "F", "Bb", "Eb", "Ab", "Db", "Gb", "Cb":
		return true
	default:
		return false
	}
}

func isChordLine(line string) bool {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return false
	}
	hasChord := false
	for _, token := range tokens {
		if isLineSeparator(token) {
			continue
		}
		if !isChordToken(token) {
			return false
		}
		hasChord = true
	}
	return hasChord
}

func transposeLine(line string, delta int, useFlats bool) string {
	parts := splitRe.Split(line, -1)
	spaceParts := splitRe.FindAllString(line, -1)
	var b strings.Builder
	for i, part := range parts {
		if part != "" {
			if isLineSeparator(part) {
				b.WriteString(part)
			} else {
				b.WriteString(transposeToken(part, delta, useFlats))
			}
		}
		if i < len(spaceParts) {
			b.WriteString(spaceParts[i])
		}
	}
	return b.String()
}

func transposeToken(token string, delta int, useFlats bool) string {
	if token == "N.C." || token == "N.C" || token == "NC" {
		return token
	}
	matches := chordRe.FindStringSubmatch(token)
	if matches == nil {
		return token
	}
	root := matches[1]
	quality := matches[2]
	bass := matches[3]
	out := transposeNote(root, delta, useFlats) + quality
	if bass != "" {
		out += "/" + transposeNote(bass, delta, useFlats)
	}
	return out
}

func transposeNote(note string, delta int, useFlats bool) string {
	note = normalizeNote(note)
	semitone, ok := semitones[note]
	if !ok {
		return note
	}
	value := (semitone + delta) % 12
	if useFlats {
		return flatNames[value]
	}
	return sharpNames[value]
}

func normalizeNote(note string) string {
	note = strings.TrimSpace(note)
	note = strings.ReplaceAll(note, "♯", "#")
	note = strings.ReplaceAll(note, "♭", "b")
	return note
}

func chordRoot(token string) string {
	matches := chordRe.FindStringSubmatch(token)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func isChordToken(token string) bool {
	if isLineSeparator(token) {
		return true
	}
	return chordRe.MatchString(token) || token == "N.C." || token == "N.C" || token == "NC"
}

func isLineSeparator(token string) bool {
	return token == "|" || token == "||" || token == "%" || token == "¦"
}

func fatal(err error) {
	if err == nil {
		return
	}
	if err.Error() == usageText {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
