package internal_test

import (
	"encoding/json"
	"testing"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/stretchr/testify/assert"
)

func TestLoadMapValues(t *testing.T) {
	// Example JSON data.
	jsonData := `{
		"a": {
			"b": [
				"c",
				{
					"d": "e",
					"f": ["g", "h", "i"]
				},
				"f"
			],
			"g": null
		},
		"h": 123,
		"i": [4,5,6]
	}`

	// Unmarshal the JSON data into a Go data structure.
	var data any
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err)
	}

	// Test cases.
	testCases := []struct {
		path string
		want any
	}{
		{"a.b[0]", "c"},
		{"a.b[1].d", "e"},
		{"a.b[2]", "f"},
		{"a.b[1].f[0]", "g"},
		{"a.b[1].f[1]", "h"},
		{"a.b[1].f[2]", "i"},
		{"h", 123.0}, // JSON unmarshals numbers to float64
		{"i[0]", 4.0},
		{"i[1]", 5.0},
		{"i[2]", 6.0},
		{"a.g", nil},
		{"a.b[3]", nil},
		{"a.c", nil},
		{"a.b[1].x", nil},
		{"a.b[1].f[3]", nil},
		{"i[3]", nil},
	}

	// Run the test cases.
	var paths []string
	for _, tc := range testCases {
		paths = append(paths, tc.path)
	}

	is := assert.New(t)
	m := internal.LoadMapValues(data, paths...)
	for _, tc := range testCases {
		is.Equal(tc.want, m[tc.path], tc.path)
	}
}

func TestDeleteMapValues(t *testing.T) {
	// Example JSON data.
	jsonData := `{
		"a": {
			"b": [
				"c",
				{
					"d": "e",
					"f": ["g", "h", "i"]
				},
				"f"
			],
			"g": null
		},
		"h": 123,
		"i": [4,5,6]
	}`

	testCases := []struct {
		path    string
		deleted bool
	}{
		{"a.b[0]", true},
		{"a.b[1].d", true},
		{"a.b[2]", true},
		{"a.b[1].f[0]", true},
		{"a.b[1].f[1]", true},
		{"a.b[1].f[2]", true},
		{"h", true},
		{"i[0]", true},
		{"i[1]", true},
		{"i[2]", true},
		{"a.g", false},
		{"a.b[3]", false},
		{"a.c", false},
		{"a.b[1].x", false},
		{"a.b[1].f[3]", false},
		{"i[3]", false},
	}

	is := assert.New(t)
	// Run the test cases.
	for _, tc := range testCases {
		// Create a copy of the original data for deletion test.
		var a any
		err := json.Unmarshal([]byte(jsonData), &a)
		if err != nil {
			t.Fatalf("Error unmarshaling JSON for deletion test: %v", err)
		}

		b, err := internal.DeleteMapValues(a, tc.path)
		is.Nil(err)

		ma := internal.LoadMapValues(a, tc.path)
		mb := internal.LoadMapValues(b, tc.path)
		if tc.deleted {
			is.NotNil(ma[tc.path], tc.path)
			is.Nil(mb[tc.path], tc.path)
		} else {
			is.Nil(ma[tc.path], tc.path)
			is.Nil(mb[tc.path], tc.path)
		}
	}
}
