package objects

import (
	"errors"
	"io"

	"bytes"

	uuid "github.com/satori/go.uuid"
)

func EmptyBlob() *Blob {
	return &Blob{
		Meta: Meta{Meta: make(map[string]string)},
	}
}

type Blob struct {
	ID   uuid.UUID
	Size int64
	Meta Meta

	Data []byte
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
		ID:   o.ID(),
		Size: o.Size(),
		Meta: o.Meta(),
	}

	return obj, obj.Decode(o)
}

func (b *Blob) Decode(o EncodedObject) error {
	if o.Type() != BlobObject {
		return errors.New("unsupported object type")
	}

	b.ID = o.ID()
	b.Meta = o.Meta()
	buf := bytes.NewBuffer(make([]byte, 0, o.Size()))
	r, err := o.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(buf, r)
	if err != nil {
		return err
	}
	b.Data = buf.Bytes()

	return err
}

func (b *Blob) Encode(o EncodedObject) error {
	o.SetType(BlobObject)
	o.SetMeta(b.Meta)
	w, err := o.Writer()
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, bytes.NewReader(b.Data))

	return err
}
