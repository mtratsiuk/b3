package utils

import (
	"testing"
)

func TestStripHtml(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<h1 id=\"fifth-example-post\">Fifth example post</h1>", "Fifth example post"},
		{"<p>Another text.\nVery <strong>important</strong> obviously.</p>", "Another text. Very important obviously."},
		{"<div><b>Bold</b> and <i>italic</i></div>", "Bold and italic"},
		{"Text without HTML", "Text without HTML"},
		{"<h1>Heading</h1>\n<p>Paragraph</p>", "Heading Paragraph"},
		{"<a href=\"https://example.com\">Link</a>", "Link"},
		{"<ul><li>Item 1</li><li>Item 2</li></ul>", "Item 1 Item 2"},
		{"<img src=\"image.jpg\" alt=\"Image\"/>", ""},
	}

	for idx, test := range tests {
		result := StripHtml(test.input)
		if result != test.expected {
			t.Errorf("%v) StripHtml('%s'): expected '%s' but got '%s'", idx, test.input, test.expected, result)
		}
	}
}
