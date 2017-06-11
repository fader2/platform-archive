package objects

import (
	"errors"
	"io"

	"bytes"

	uuid "github.com/satori/go.uuid"
)

type Blob struct {
	ID          uuid.UUID
	Size        int64
	ContentType string // mime

	Data []byte

	obj EncodedObject
}

func GetBlob(s Storer, id uuid.UUID) (*Blob, error) {
	o, err := s.EncodedObject(BlobObject, id)
	if err != nil {
		return nil, err
	}

	return DecodeBlob(o)
}

func SetBlob(s Storer, b *Blob) (uuid.UUID, error) {
	obj := s.NewEncodedObject(b.ID)
	if err := b.Encode(obj); err != nil {
		return b.ID, err
	}
	return s.SetEncodedObject(obj)
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

func (b *Blob) Decode(o EncodedObject) error {
	if o.Type() != BlobObject {
		return errors.New("unsupported object type")
	}

	b.ID = o.ID()
	b.ContentType = o.ContentType()
	buf := bytes.NewBuffer(b.Data)
	r, err := o.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(buf, r)

	return err
}

func (b *Blob) Encode(o EncodedObject) error {
	o.SetType(BlobObject)
	o.SetContentType(b.ContentType)
	w, err := o.Writer()
	if err != nil {
		return err
	}
	defer w.Close()
	buf := bytes.NewBuffer(b.Data)
	defer w.Close()
	_, err = io.Copy(w, buf)

	return err
}

func (b *Blob) Reader() (io.ReadCloser, error) {
	return b.obj.Reader()
}

func (b *Blob) Writer() (io.WriteCloser, error) {
	return b.obj.Writer()
}
