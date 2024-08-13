package textdump_test

import (
	"io"
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

func TestSnapshot(t *testing.T) {
	rw := &mockReadWriter{}
	if err := textdump.Snapshot(rw, []byte("hello world")); err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(rw)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "hello world", string(b); want != got {
		t.Fatalf("want %s, got %s", want, got)
	}

	err = textdump.Snapshot(rw, []byte("hello"))
	if err == nil {
		t.Error("expected error")
	}
	t.Log(err)
}

type mockReadWriter struct {
	data      []byte
	readIndex int64
}

func (m *mockReadWriter) Read(b []byte) (n int, err error) {
	if m.readIndex >= int64(len(m.data)) {
		err = io.EOF
		return
	}

	n = copy(b, m.data[m.readIndex:])
	m.readIndex += int64(n)
	return
}

func (m *mockReadWriter) Write(b []byte) (n int, err error) {
	if len(m.data) > 0 {
		return 0, nil
	}

	m.data = b
	return len(b), nil
}
