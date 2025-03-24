package utils

import (
	"testing"
)

func TestTrimText(t *testing.T) {
	breakAt := 16

	tests := []struct {
		input    string
		expected string
	}{
		{"One two three", "One two three"},
		{"One two three fo", "One two three fo"},
		{"One two three four five six seven", "One two three..."},
		{"One two three fo ur", "One two three fo..."},
		{"Onetwothreefourfive", "Onetwothreefourf..."},
		{"One two three       ", "One two three..."},
		{"                    ", "..."},
	}

	for idx, test := range tests {
		result := TrimText(test.input, breakAt)
		if result != test.expected {
			t.Errorf("%v) TrimText('%v', %v): expected '%v' but got '%v'", idx, test.input, breakAt, test.expected, result)
		}
	}
}
