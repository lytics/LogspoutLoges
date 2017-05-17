package logspoutloges

import "strings"

func escapeNewlines(str string) string {
	return strings.Replace(str, "\n", "\\n", -1)
}

func encodeNewlines(str string) string {
	return strings.Replace(str, "\\n", "\n", -1)
}
