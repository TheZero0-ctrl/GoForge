package naming

import (
	"fmt"
	"strings"
	"unicode"
)

type ResourceNames struct {
	Input          string
	Singular       string
	Plural         string
	SingularPascal string
	PluralPascal   string
	SingularCamel  string
	PluralCamel    string
}

var irregularSingularToPlural = map[string]string{
	"person": "people",
	"man":    "men",
	"woman":  "women",
	"child":  "children",
	"mouse":  "mice",
}

var irregularPluralToSingular = reverseMap(irregularSingularToPlural)

func init() {
	irregularPluralToSingular["movies"] = "movie"
}

func NormalizeResourceName(raw string) (ResourceNames, error) {
	snake := ToSnake(raw)
	if snake == "" {
		return ResourceNames{}, fmt.Errorf("resource name cannot be empty")
	}

	if !IsSnakeIdentifier(snake) {
		return ResourceNames{}, fmt.Errorf("resource name %q is invalid; use letters, numbers, and underscores only", raw)
	}

	singular := Singularize(snake)
	plural := Pluralize(singular)

	return ResourceNames{
		Input:          raw,
		Singular:       singular,
		Plural:         plural,
		SingularPascal: ToPascal(singular),
		PluralPascal:   ToPascal(plural),
		SingularCamel:  ToLowerCamel(singular),
		PluralCamel:    ToLowerCamel(plural),
	}, nil
}

func Pluralize(word string) string {
	w := ToSnake(word)
	if w == "" {
		return ""
	}

	if irregular, ok := irregularSingularToPlural[w]; ok {
		return irregular
	}

	if strings.HasSuffix(w, "y") && len(w) > 1 && !isVowel(rune(w[len(w)-2])) {
		return w[:len(w)-1] + "ies"
	}

	if hasAnySuffix(w, "s", "x", "z", "ch", "sh") {
		return w + "es"
	}

	return w + "s"
}

func Singularize(word string) string {
	w := ToSnake(word)
	if w == "" {
		return ""
	}

	if irregular, ok := irregularPluralToSingular[w]; ok {
		return irregular
	}

	candidates := make([]string, 0, 3)
	if strings.HasSuffix(w, "ies") && len(w) > 3 {
		candidates = append(candidates, w[:len(w)-3]+"y")
	}

	if hasAnySuffix(w, "ses", "xes", "zes", "ches", "shes") {
		candidates = append(candidates, w[:len(w)-2])
	}

	if strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && len(w) > 1 {
		candidates = append(candidates, w[:len(w)-1])
	}

	for _, candidate := range candidates {
		if Pluralize(candidate) == w {
			return candidate
		}
	}

	return w
}

func ToSnake(input string) string {
	parts := splitWords(input)
	return strings.Join(parts, "_")
}

func ToPascal(input string) string {
	parts := splitWords(input)
	if len(parts) == 0 {
		return ""
	}

	for i := range parts {
		runes := []rune(parts[i])
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	return strings.Join(parts, "")
}

func ToLowerCamel(input string) string {
	pascal := ToPascal(input)
	if pascal == "" {
		return ""
	}

	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func IsSnakeIdentifier(value string) bool {
	if value == "" {
		return false
	}

	runes := []rune(value)
	if !unicode.IsLower(runes[0]) {
		return false
	}

	for _, r := range runes[1:] {
		if unicode.IsLower(r) || unicode.IsDigit(r) || r == '_' {
			continue
		}
		return false
	}

	return true
}

func splitWords(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	runes := []rune(input)
	parts := make([]string, 0, 4)
	current := make([]rune, 0, len(runes))

	flush := func() {
		if len(current) == 0 {
			return
		}
		parts = append(parts, strings.ToLower(string(current)))
		current = current[:0]
	}

	for i, r := range runes {
		if isSeparator(r) {
			flush()
			continue
		}

		if unicode.IsUpper(r) && len(current) > 0 {
			prev := runes[i-1]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextIsLower) {
				flush()
			}
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, unicode.ToLower(r))
			continue
		}

		flush()
	}

	flush()
	return parts
}

func isSeparator(r rune) bool {
	return r == '_' || r == '-' || unicode.IsSpace(r) || r == '.' || r == '/'
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

func reverseMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[value] = key
	}
	return out
}
