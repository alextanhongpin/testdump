package snapshot

import "io"

type encoder interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte) (any, error)
}

type comparer interface {
	Compare(a, b any) error
}

func Snapshot(rw io.ReadWriter, enc encoder, cmp comparer, v any) error {
	b, err := enc.Marshal(v)
	if err != nil {
		return err
	}

	n, err := rw.Write(b)
	if err != nil {
		return err
	}
	if n != 0 {
		return nil
	}

	a, err := io.ReadAll(rw)
	if err != nil {
		return err
	}

	snap, err := enc.Unmarshal(a)
	if err != nil {
		return err
	}

	recv, err := enc.Unmarshal(b)
	if err != nil {
		return err
	}

	return cmp.Compare(snap, recv)
}
