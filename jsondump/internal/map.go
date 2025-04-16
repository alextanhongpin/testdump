package internal

import (
	"fmt"
	"strconv"
	"strings"
)

// Define custom error types for better error handling.
type KeyNotFoundError struct {
	Path    string
	Segment string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("path '%s' is invalid: key '%s' not found", e.Path, e.Segment)
}

type OutOfBoundsError struct {
	Path    string
	Index   int
	Length  int
	Segment string
}

func (e *OutOfBoundsError) Error() string {
	return fmt.Sprintf("path '%s' is invalid: index %d out of bounds for slice of length %d at segment '%s'", e.Path, e.Index, e.Length, e.Segment)
}

type InvalidTypeError struct {
	Path    string
	Segment string
	Type    string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("path '%s' is invalid: expected map or slice at segment '%s', but got %s", e.Path, e.Segment, e.Type)
}

// GetValue retrieves a value from a JSON object, supporting deep nesting and slice indexing.
//
// The path argument is a string that specifies the path to the desired value within the JSON object.
// The path is a dot-separated string, where each segment represents a key in the JSON object.
// If a segment is an integer, it is interpreted as an index into a JSON array.
//
// For example, given the following JSON object:
//
//	{
//		"a": {
//			"b": [
//				"c",
//				{
//					"d": "e"
//				},
//				"f"
//			]
//		}
//	}
//
// The following calls to GetValue would return the following values:
//
//	GetValue(data, "a.b.0") // returns "c"
//	GetValue(data, "a.b.1.d") // returns "e"
//	GetValue(data, "a.b.2") // returns "f"
//
// If the path is invalid, or if the value at the specified path is not of the expected type,
// GetValue returns nil and an error.
//
// data: The JSON object to query.  It is of type any, but is expected to be a
//
//	map[string]any or a []any.
//
// path: The path to the value to retrieve.  For example, "a.b.c" or "a.b.1.d".
//
// Returns:
// The value at the specified path, or nil if the path is invalid.
// An error if the path is invalid or the value is not of the expected type.
func GetValue(data any, path string) (any, error) {
	// Split the path into segments.
	segments := strings.Split(path, ".")

	// Iterate over the segments, traversing the JSON object.
	current := data
	for i, segment := range segments {
		// Check if the current value is nil.
		if current == nil {
			return nil, fmt.Errorf("path '%s' is invalid: nil value at segment '%s'", path, strings.Join(segments[:i], "."))
		}

		// Try to interpret the segment as an integer.
		index, err := strconv.Atoi(segment)
		if err == nil {
			// The segment is an integer, so treat the current value as a slice.
			slice, ok := current.([]any)
			if !ok {
				return nil, &InvalidTypeError{
					Path:    path,
					Segment: strings.Join(segments[:i+1], "."),
					Type:    fmt.Sprintf("%T", current),
				}
			}

			// Check if the index is within the bounds of the slice.
			if index < 0 || index >= len(slice) {
				return nil, &OutOfBoundsError{
					Path:    path,
					Index:   index,
					Length:  len(slice),
					Segment: strings.Join(segments[:i+1], "."),
				}
			}

			// Get the value at the specified index.
			current = slice[index]
		} else {
			// The segment is not an integer, so treat the current value as a map.
			mapping, ok := current.(map[string]any)
			if !ok {
				return nil, &InvalidTypeError{
					Path:    path,
					Segment: strings.Join(segments[:i+1], "."),
					Type:    fmt.Sprintf("%T", current),
				}
			}

			// Check if the key exists in the map.
			value, ok := mapping[segment]
			if !ok {
				return nil, &KeyNotFoundError{
					Path:    path,
					Segment: segment,
				}
			}

			// Get the value for the specified key.
			current = value
		}
	}

	// Return the final value.
	return current, nil
}

// DeleteValue deletes a value from a JSON object at the given path.
//
// The path argument is a string that specifies the path to the value to delete within the JSON object.
// The path is a dot-separated string, where each segment represents a key in the JSON object.
// If a segment is an integer, it is interpreted as an index into a JSON array.
//
// For example, given the following JSON object:
//
//	{
//		"a": {
//			"b": [
//				"c",
//				{
//					"d": "e"
//				},
//				"f"
//			]
//		}
//	}
//
// The following calls to DeleteValue would delete the following values:
//
//	DeleteValue(data, "a.b.0") // deletes "c"
//	DeleteValue(data, "a.b.1.d") // deletes "e"
//	DeleteValue(data, "a.b.2") // deletes "f"
//
// If the path is invalid, or if the value at the specified path is not of the expected type,
// DeleteValue returns an error.
//
// data: The JSON object to modify.  It is of type any, but is expected to be a
//
//	map[string]any or a []any.
//
// path: The path to the value to delete.  For example, "a.b.c" or "a.b.1.d".
//
// Returns:
// An error if the path is invalid.
func DeleteValue(data any, path string) error {
	// Split the path into segments.
	segments := strings.Split(path, ".")

	// Iterate over the segments, traversing the JSON object.
	current := data
	for i, segment := range segments {
		// Check if the current value is nil.
		if current == nil {
			return fmt.Errorf("path '%s' is invalid: nil value at segment '%s'", path, strings.Join(segments[:i], "."))
		}

		// Try to interpret the segment as an integer.
		index, err := strconv.Atoi(segment)
		if err == nil {
			// The segment is an integer, so treat the current value as a slice.
			slice, ok := current.([]any)
			if !ok {
				return &InvalidTypeError{
					Path:    path,
					Segment: strings.Join(segments[:i+1], "."),
					Type:    fmt.Sprintf("%T", current),
				}
			}

			// Check if the index is within the bounds of the slice.
			if index < 0 || index >= len(slice) {
				return &OutOfBoundsError{
					Path:    path,
					Index:   index,
					Length:  len(slice),
					Segment: strings.Join(segments[:i+1], "."),
				}
			}

			if i == len(segments)-1 {
				// If this is the last segment, delete the element from the slice.
				currentSlice := current.([]any)
				currentSlice = append(currentSlice[:index], currentSlice[index+1:]...)
				return nil
			}

			// Get the value at the specified index.
			current = slice[index]
		} else {
			// The segment is not an integer, so treat the current value as a map.
			mapping, ok := current.(map[string]any)
			if !ok {
				return &InvalidTypeError{
					Path:    path,
					Segment: strings.Join(segments[:i+1], "."),
					Type:    fmt.Sprintf("%T", current),
				}
			}

			if i == len(segments)-1 {
				// If this is the last segment, delete the key from the map.
				delete(mapping, segment)
				return nil
			}

			// Check if the key exists in the map.
			value, ok := mapping[segment]
			if !ok {
				return &KeyNotFoundError{
					Path:    path,
					Segment: segment,
				}
			}

			// Get the value for the specified key.
			current = value
		}
	}
	return nil
}
