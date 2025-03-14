package service

import (
	"regexp"
	"strings"
)

var placeholderRE = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// RenderTemplate replaces {{key}} placeholders with values from vars.
// Unknown placeholders are left unchanged.
func RenderTemplate(input string, vars map[string]string) string {
	if len(vars) == 0 {
		return input
	}
	return placeholderRE.ReplaceAllStringFunc(input, func(match string) string {
		sub := placeholderRE.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := strings.TrimSpace(sub[1])
		if val, ok := vars[key]; ok {
			return val
		}
		return match
	})
}
