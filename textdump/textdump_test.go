package textdump_test

import (
	"testing"

	"github.com/alextanhongpin/dump/textdump"
)

func TestDump(t *testing.T) {
	textdump.Dump(t, []byte("hello world"))
}
