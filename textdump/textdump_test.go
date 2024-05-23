package textdump_test

import (
	"testing"

	"github.com/alextanhongpin/testdump/textdump"
)

func TestDump(t *testing.T) {
	textdump.Dump(t, []byte("hello world"))
}
