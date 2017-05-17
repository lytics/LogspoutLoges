package logspoutloges

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestNewlineEscape(t *testing.T) {
	tstr := "hihi\nhihi"
	e := "hihi\\nhihi"

	ret := escapeNewlines(tstr)
	assert.Tf(t, ret == e, "Escaped newlines incorrect!")
}
func TestNewlineEscape2(t *testing.T) {
	tstr := "hihi\nhihi\n"
	e := "hihi\\nhihi\\n"

	ret := escapeNewlines(tstr)
	assert.Tf(t, ret == e, "Escaped newlines incorrect!")
}

func TestNewlineEncode(t *testing.T) {
	e := "hihi\nhihi"
	tstr := "hihi\\nhihi"

	ret := encodeNewlines(tstr)
	assert.Tf(t, ret == e, "Encoded newlines incorrect!")
}

func TestNewlineEncode2(t *testing.T) {
	e := "hihi\nhihi\n"
	tstr := "hihi\\nhihi\\n"

	ret := encodeNewlines(tstr)
	assert.Tf(t, ret == e, "Encoded newlines incorrect!")
}

func TestNewlineEncode3(t *testing.T) {
	e := "\n\nhihi\n"
	tstr := "\\n\\nhihi\\n"

	ret := encodeNewlines(tstr)
	assert.Tf(t, ret == e, "Encoded newlines incorrect!")
}
