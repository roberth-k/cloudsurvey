package core

import (
	"io"
	"os"
)

type StringLineWriter interface {
	WriteStringLine(s string) (int, error)
}

type FileStringLineWriter struct {
	*os.File
}

func (f *FileStringLineWriter) WriteStringLine(s string) (int, error) {
	n1, err := f.WriteString(s)
	if err != nil {
		return n1, err
	}

	n2, err := f.WriteString("\n")

	return n1 + n2, err
}

type WriterStringLineWriter struct {
	io.Writer
}

func (f WriterStringLineWriter) WriteStringLine(s string) (int, error) {
	n1, err := f.Write([]byte(s))
	if err != nil {
		return n1, err
	}

	n2, err := f.Write([]byte{'\n'})

	return n1 + n2, err
}
