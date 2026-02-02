package sqlformat

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

func Format(stmt string) (string, error) {
	b, err := sqlformat(stmt)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// equals to $ echo 'select 1' | python3 -m sqlparse -r -
func sqlformat(stmt string) ([]byte, error) {
	r, w := io.Pipe()
	defer r.Close()

	echo := exec.Command("echo", stmt)
	sqlparse := exec.Command("uvx", "--from", "sqlparse",
		"sqlformat",
		"--reindent",           // Reindent statements
		"--indent_after_first", // Indent after first line of statement
		"--keywords", "upper",  // Change case of keywords - "upper", "lower" or "capitalize"
		"--strip-comments", // Remove comments
		"-")
	echo.Stdout = w
	sqlparse.Stdin = r
	defer w.Close()

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	sqlparse.Stderr = &stderr
	sqlparse.Stdout = &stdout

	if err := echo.Start(); err != nil {
		return nil, err
	}

	if err := sqlparse.Start(); err != nil {
		return nil, err
	}

	if err := echo.Wait(); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	if err := sqlparse.Wait(); err != nil {
		return nil, err
	}
	if b := stderr.Bytes(); len(b) > 0 {
		return nil, errors.New(string(b))
	}
	return stdout.Bytes(), nil
}
