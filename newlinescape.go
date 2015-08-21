package logspoutloges

import "strings"

func EscapeNewlines(str string) string {
	return strings.Replace(str, "\n", "\\n", -1)
}

func EncodeNewlines(str string) string {
	return strings.Replace(str, "\\n", "\n", -1)
}
