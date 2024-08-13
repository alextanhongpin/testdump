package snapshot

import "io"

type encoder interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte) (any, error)
}

func Snapshot(w io.Writer, r io.Reader, enc encoder, v any, compare func(a, b any) error) error {
	b, err := enc.Marshal(v)
	if err != nil {
		return err
	}

	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	c, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	recv, err := enc.Unmarshal(b)
	if err != nil {
		return err
	}

	snap, err := enc.Unmarshal(c)
	if err != nil {
		return err
	}

	return compare(snap, recv)
}
