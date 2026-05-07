package tab

import (
	"fmt"
	"regexp"
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

func FetchTab(id int64) (ultimateguitar.TabResult, error) {
	s := ultimateguitar.New()
	return s.GetTabByID(id)
}

func TabSourceKey(tab ultimateguitar.TabResult, content string) string {
	sourceKey := inferKey(StripUGMarkup(content))
	if sourceKey == "" {
		sourceKey = strings.TrimSpace(tab.TonalityName)
	}
	if sourceKey == "" {
		sourceKey = strings.TrimSpace(tab.Recording.TonalityName)
	}
	return sourceKey
}

func Slugify(value string) string {
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

func SafeComponent(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("/", "-", "\\", "-", ":", "-", "\x00", "").Replace(value)
	if value == "" {
		return "unknown"
	}
	return value
}

func StripLeadingMetadata(content string) string {
	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[start]), "[") {
		start++
	}
	return strings.Join(lines[start:], "\n")
}

func StripUGMarkup(content string) string {
	return markupReplacer.Replace(content)
}

func Transpose(content, sourceKey, targetKey string) (string, error) {
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

func TransposeMarkup(content, sourceKey, targetKey string) (string, error) {
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
