package boltdb

import (
	"errors"
	"io"
)

// Reader

func newReader(r io.Reader) *objReader {
	return &objReader{
		raw: r,
	}
}

type objReader struct {
	raw io.Reader
}

func (r *objReader) Header() (t ObjectType, ct string, err error) {
	tb, err := r.readUntil(' ')
	if err != nil {
		return
	}
	ctb, err := r.readUntil(' ')
	if err != nil {
		return
	}
	t = ObjectType(string(tb))
	ct = string(ctb)
	return
}

func (r *objReader) readUntil(delim byte) ([]byte, error) {
	var buf [1]byte
	value := make([]byte, 0, 255)
	for {
		if n, err := r.raw.Read(buf[:]); err != nil && (err != io.EOF || n == 0) {
			if err == io.EOF {
				return nil, errors.New("err header")
			}
			return nil, err
		}

		if buf[0] == delim {
			return value, nil
		}

		value = append(value, buf[0])
	}
}

func (w *objReader) Read(p []byte) (n int, err error) {
	return w.raw.Read(p)
}

func (w *objReader) Close() error {
	return nil
}

// Writer

func newWriter(w io.Writer) *objWriter {
	return &objWriter{
		raw: w,
	}
}

type objWriter struct {
	raw io.Writer
}

func (w *objWriter) WriteHeader(
	t ObjectType,
	ct string, // content type (recommended used mime type)
) (err error) {
	b := t.Bytes()
	b = append(b, ' ')
	b = append(b, ct...)
	b = append(b, ' ')

	_, err = w.Write(b)
	return err
}

func (w *objWriter) Write(p []byte) (n int, err error) {
	return w.raw.Write(p)
}

func (w *objWriter) Close() error {
	return nil
}
