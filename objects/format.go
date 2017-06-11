package objects

import (
	"errors"
	"io"
)

// Reader

func NewReader(r io.Reader) *ObjectReader {
	return &ObjectReader{
		raw: r,
	}
}

type ObjectReader struct {
	raw io.Reader
}

func (r *ObjectReader) Header() (t ObjectType, ct string, err error) {
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

func (r *ObjectReader) readUntil(delim byte) ([]byte, error) {
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

func (w *ObjectReader) Read(p []byte) (n int, err error) {
	return w.raw.Read(p)
}

func (w *ObjectReader) Close() error {
	return nil
}

// Writer

func NewWriter(w io.Writer) *ObjectWriter {
	return &ObjectWriter{
		raw: w,
	}
}

type ObjectWriter struct {
	raw io.Writer
}

func (w *ObjectWriter) WriteHeader(
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

func (w *ObjectWriter) Write(p []byte) (n int, err error) {
	return w.raw.Write(p)
}

func (w *ObjectWriter) Close() error {
	return nil
}
