package cmd

import (
	"fmt"
	"testing"
)

// The wanted values are what notes.py's parse() answers for the same text
// (checked against /opt/agents/bin/notes.py); no input may panic.
func TestParseNoteMalformed(t *testing.T) {
	for _, tc := range []struct {
		text, meta, body string
	}{
		{"", "map[]", ""},
		{"---", "map[]", "---"},
		{"---\n", "map[]", "---\n"},
		{"---\n---", "map[]", ""},
		{"---\n---\n", "map[]", "---\n---\n"},
		{"---\n\n---\n", "map[]", ""},
		{"---\ntitle: A\n---", "map[title:A]", ""},
		{"---\ntitle: A\n---\n", "map[title:A]", ""},
		{"---\ntitle: A\nbody without close", "map[]", "---\ntitle: A\nbody without close"},
		{"---\r\ntitle: A\r\n---\r\nbody\r\n", "map[]", "---\r\ntitle: A\r\n---\r\nbody\r\n"},
		{"---\ntitle: A\r\n---\nbody\r\n", "map[title:A]", "body\r\n"},
		{"no front\n---\n", "map[]", "no front\n---\n"},
		{"---\nTitle: x\nbad line\nk: v: w\n---\nb", "map[k:v: w]", "b"},
	} {
		meta, body := parseNote(tc.text)
		if got := fmt.Sprint(meta); got != tc.meta || body != tc.body {
			t.Errorf("parseNote(%q) = %s, %q; want %s, %q", tc.text, got, body, tc.meta, tc.body)
		}
	}
}
