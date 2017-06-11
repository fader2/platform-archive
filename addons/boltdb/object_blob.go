package boltdb

import (
	"io"

	uuid "github.com/satori/go.uuid"
)

type Blob struct {
	ID          uuid.UUID
	Size        int64
	ContentType string // mime

	obj EncodedObject
}

func GetBlob(s Storer, id uuid.UUID) (*Blob, error) {
	o, err := s.EncodedObject(BlobObject, id)
	if err != nil {
		return nil, err
	}

	return DecodeBlob(o)
}

func DecodeBlob(o EncodedObject) (*Blob, error) {
	obj := &Blob{
		ID:          o.ID(),
		Size:        o.Size(),
		ContentType: o.ContentType(),
		obj:         o,
	}

	return obj, nil
}

func (b *Blob) Reader() (io.ReadCloser, error) {
	return b.obj.Reader()
}

func (b *Blob) Writer() (io.WriteCloser, error) {
	return b.obj.Writer()
}
