package file

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

var _ io.ReadWriteCloser = (*File)(nil)

type File struct {
	exists    bool
	f         *os.File
	name      string
	overwrite bool
}

func New(name string, overwrite bool) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err == nil {
		return &File{
			f:         f,
			name:      name,
			overwrite: overwrite,
		}, nil
	}

	if !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	f, err = os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &File{
		exists:    true,
		f:         f,
		name:      name,
		overwrite: overwrite,
	}, nil
}

func (f *File) Write(b []byte) (int, error) {
	if f.exists {
		if f.overwrite {
			// We need to truncate the file content.
			f, err := os.OpenFile(f.name, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			return f.Write(b)
		}

		return 0, nil
	}

	return f.f.Write(b)
}

func (f *File) Read(b []byte) (int, error) {
	return f.f.Read(b)
}

func (f *File) Close() error {
	return f.f.Close()
}
