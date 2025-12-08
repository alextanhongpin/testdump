package internal_test

import (
	"testing"

	"github.com/alextanhongpin/testdump/jsondump/internal"
)

func TestJSON(t *testing.T) {
	b := []byte(`{"name": "alice", "hobbies": ["swimming"], "friends": [{"name": "bob"}]}`)
	res, err := internal.ReplaceJSON(b, map[string]any{
		"name":           "bob",
		"hobbies":        "dancing", // Doesn't work.
		"friends/0/name": "bobys",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}
