// Updated utils.ExtractJSON function
package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ExtractJSON(input string, out any) error {
	// Clean Markdown code blocks first
	mdRegex := regexp.MustCompile("(?i)(\\s*```json|\\s*```)")
	cleaned := mdRegex.ReplaceAllString(input, "")

	// Preserve valid escape sequences
	cleaned = strings.ReplaceAll(cleaned, `\\n`, "\n")
	cleaned = strings.ReplaceAll(cleaned, `\n`, "\n")
	cleaned = strings.ReplaceAll(cleaned, `\t`, "\t")
	cleaned = strings.ReplaceAll(cleaned, `\\"`, `"`)

	if json.Unmarshal([]byte(cleaned), out) == nil {
		return nil
	}

	jsonRegex := regexp.MustCompile(`(?s)\{(\{|[^{}])*\}|\[(\[|[^\[\]])*\]`)
	matches := jsonRegex.FindAllString(cleaned, -1)

	for i := len(matches) - 1; i >= 0; i-- {
		candidate := matches[i]
		repaired := repairJSON(candidate)

		if err := json.Unmarshal([]byte(repaired), out); err == nil {
			return nil
		}
	}

	return json.NewDecoder(strings.NewReader("{}")).Decode(out)
}

// repairJSON remains mostly the same with improved regex
func repairJSON(candidate string) string {
	// Fix common formatting issues
	repaired := strings.ReplaceAll(candidate, `'`, `"`)

	// Add quotes around unquoted keys
	repaired = regexp.MustCompile(`([{,]\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`).ReplaceAllString(repaired, `$1"$2":`)

	// Add quotes around unquoted string values
	repaired = regexp.MustCompile(`:\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*([,}])`).ReplaceAllString(repaired, `:"$1"$2`)

	// Remove trailing commas
	repaired = regexp.MustCompile(`,\s*([}\]])`).ReplaceAllString(repaired, "$1")

	// Balance braces/brackets if needed
	openBraces := strings.Count(repaired, "{") - strings.Count(repaired, "}")
	openBrackets := strings.Count(repaired, "[") - strings.Count(repaired, "]")

	if openBraces > 0 {
		repaired += strings.Repeat("}", openBraces)
	} else if openBraces < 0 {
		repaired = strings.Repeat("{", -openBraces) + repaired
	}

	if openBrackets > 0 {
		repaired += strings.Repeat("]", openBrackets)
	} else if openBrackets < 0 {
		repaired = strings.Repeat("[", -openBrackets) + repaired
	}

	return repaired
}
