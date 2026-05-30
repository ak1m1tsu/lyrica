package template

import (
	"testing"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		input float64
		want  string
	}{
		{0, "0:00"},
		{60.0, "1:00"},
		{61.0, "1:01"},
		{214.0, "3:34"},
		{3600.0, "60:00"},
		{90.5, "1:30"},
	}
	for _, c := range cases {
		got := formatDuration(c.input)
		if got != c.want {
			t.Errorf("formatDuration(%v) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestExtractTimestamp(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"[02:30.15] Hello", "[02:30.15]"},
		{"[00:00.00] intro", "[00:00.00]"},
		{"no timestamp", ""},
		{"", ""},
		{"[missing close", ""},
		{"[] empty", ""},
	}
	for _, c := range cases {
		got := extractTimestamp(c.input)
		if got != c.want {
			t.Errorf("extractTimestamp(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestExtractLyricText(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"[02:30.15] Hello world", "Hello world"},
		{"[00:00.00]   leading spaces", "leading spaces"},
		{"plain line", "plain line"},
		{"", ""},
		{"[missing close", "[missing close"},
		{"[02:30.15]", "[02:30.15]"},
	}
	for _, c := range cases {
		got := extractLyricText(c.input)
		if got != c.want {
			t.Errorf("extractLyricText(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
