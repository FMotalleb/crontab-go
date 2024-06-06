package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type writer struct {
	stdout bool
	file   *os.File
}

func (w *writer) Write(b []byte) (n int, e error) {
	reader := bytes.NewReader(b)
	var written int64
	var err error
	if w.stdout {
		written, err = io.Copy(os.Stdout, reader)
	}
	n = int(written)
	if err != nil {
		return n, err
	}

	if w.file != nil {
		wf, er := w.file.Write(b)
		if wf != n {
			return n, fmt.Errorf("wrote two different amount of log into stdout(%d) and file(%d)", written, wf)
		}
		if er != nil {
			err = er
		}
	}
	return n, err
}
