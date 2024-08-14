package textdump_test

import (
	"testing"

	"github.com/alextanhongpin/testdump/textdump"
)

func TestDump(t *testing.T) {
	textdump.Dump(t, []byte("hello world"))
}

func TestDumpFile(t *testing.T) {
	textdump.Dump(t, []byte("foo"), textdump.File("foo"))
	textdump.Dump(t, []byte("bar"), textdump.File("bar"))
}
