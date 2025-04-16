package internal_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/alextanhongpin/testdump/jsondump/internal"
)

// Import the main package to access its functions.

func TestGetValue(t *testing.T) {
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
		path     string
		expected any
		err      error // Use nil for no error, otherwise, check for non-nil
	}{
		{"a.b.0", "c", nil},
		{"a.b.1.d", "e", nil},
		{"a.b.2", "f", nil},
		{"a.b.1.f.0", "g", nil},
		{"a.b.1.f.1", "h", nil},
		{"a.b.1.f.2", "i", nil},
		{"h", 123.0, nil}, // JSON unmarshals numbers to float64
		{"i.0", 4.0, nil},
		{"i.1", 5.0, nil},
		{"i.2", 6.0, nil},
		{"a.g", nil, nil},
		{"a.b.3", nil, &internal.OutOfBoundsError{}},
		{"a.c", nil, &internal.KeyNotFoundError{}},
		{"a.b.1.x", nil, &internal.KeyNotFoundError{}},
		{"a.b.1.f.3", nil, &internal.OutOfBoundsError{}},
		{"i.3", nil, &internal.OutOfBoundsError{}},
	}

	// Run the test cases.
	for _, tc := range testCases {
		actual, err := internal.GetValue(data, tc.path)
		if tc.err == nil {
			if err != nil {
				t.Errorf("Test case '%s' failed: expected no error, but got %v\n", tc.path, err)
				continue
			}
			// Compare the actual value with the expected value.  Use fmt.Sprintf to
			// handle the fact that JSON numbers are unmarshaled as float64.
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", tc.expected) {
				t.Errorf("Test case '%s' failed: expected %v, but got %v\n", tc.path, tc.expected, actual)
			} else {
				t.Logf("Test case '%s' passed: got %v\n", tc.path, actual)
			}
		} else {
			if err == nil {
				t.Errorf("Test case '%s' failed: expected error '%v', but got no error\n", tc.path, tc.err)
				continue
			}
			// Check if the error is of the expected type.
			if _, ok := err.(error); !ok {
				t.Errorf("Test case '%s' failed: expected error of type '%T', but got '%T'\n", tc.path, tc.err, err)
				continue
			}
			if !strings.Contains(err.Error(), tc.err.Error()) {
				t.Errorf("Test case '%s' failed: expected error containing '%v', but got '%v'\n", tc.path, tc.err, err)
			} else {
				t.Logf("Test case '%s' passed: got expected error\n", tc.path)
			}
		}
	}
}

func TestDeleteValue(t *testing.T) {
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
		err     error
	}{
		{"a.b.0", true, nil},
		{"a.b.1.d", true, nil},
		{"a.b.2", true, nil},
		{"a.b.1.f.0", true, nil},
		{"a.b.1.f.1", true, nil},
		{"a.b.1.f.2", true, nil},
		{"h", true, nil},
		{"i.0", true, nil},
		{"i.1", true, nil},
		{"i.2", true, nil},
		{"a.g", true, nil},
		{"a.b.3", false, &internal.OutOfBoundsError{}},
		{"a.c", false, &internal.KeyNotFoundError{}},
		{"a.b.1.x", false, &internal.KeyNotFoundError{}},
		{"a.b.1.f.3", false, &internal.OutOfBoundsError{}},
		{"i.3", false, &internal.OutOfBoundsError{}},
	}

	// Run the test cases.
	for _, tc := range testCases {
		// Create a copy of the original data for deletion test.
		var deleteData any
		err := json.Unmarshal([]byte(jsonData), &deleteData)
		if err != nil {
			t.Fatalf("Error unmarshaling JSON for deletion test: %v", err)
		}

		deleteErr := internal.DeleteValue(deleteData, tc.path)

		if tc.err == nil {
			if deleteErr != nil {
				t.Errorf("Test case '%s' for DeleteValue failed: expected no error, but got %v\n", tc.path, deleteErr)
				continue
			}
			if tc.deleted {
				// Verify the value is deleted by trying to get it.
				_, getErr := internal.GetValue(deleteData, tc.path)
				if getErr == nil {
					t.Errorf("Test case '%s' for DeleteValue failed: expected value to be deleted, but it still exists\n", tc.path)
					continue
				}
				t.Logf("Test case '%s' for DeleteValue passed: value deleted\n", tc.path)
			}

		} else {
			if deleteErr == nil {
				t.Errorf("Test case '%s' for DeleteValue failed: expected error containing '%v', but got no error\n", tc.path, tc.err)
				continue
			}
			// Check if the error is of the expected type.
			if _, ok := deleteErr.(error); !ok {
				t.Errorf("Test case '%s' failed: expected error of type '%T', but got '%T'\n", tc.path, tc.err, deleteErr)
				continue
			}
			if !strings.Contains(deleteErr.Error(), tc.err.Error()) {
				t.Errorf("Test case '%s' failed: expected error containing '%v', but got '%v'\n", tc.path, tc.err, deleteErr)
			} else {
				t.Logf("Test case '%s' passed: got expected error\n", tc.path)
			}
		}
	}
}
