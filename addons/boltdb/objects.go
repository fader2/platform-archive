package boltdb

import (
	"bytes"
	"io"
	"io/ioutil"

	uuid "github.com/satori/go.uuid"
)

type ObjectType string

const (
	InvalidObject ObjectType = ""
	BlobObject    ObjectType = "blob"
)

func (t ObjectType) Bytes() []byte {
	return []byte(string(t))
}

var _ EncodedObject = (*obj)(nil)

type obj struct {
	id uuid.UUID
	t  ObjectType
	ct string // content type
	d  []byte
	sz int64
}

func (o *obj) ID() uuid.UUID {
	return o.id
}

func (o *obj) Size() int64 {
	return o.sz
}

func (o *obj) SetSize(s int64) {
	o.sz = s
}

func (o *obj) Type() ObjectType {
	return o.t
}

func (o *obj) ContentType() string {
	return o.ct
}

func (o *obj) SetContentType(ct string) {
	o.ct = ct
}

func (o *obj) SetType(t ObjectType) {
	o.t = t
}

func (o *obj) Reader() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(o.d)), nil
}

func (o *obj) Writer() (io.WriteCloser, error) {
	return o, nil
}

func (o *obj) Write(p []byte) (n int, err error) {
	o.d = append(o.d, p...)
	o.sz = int64(len(o.d))

	return len(p), nil
}

func (o *obj) Close() error {
	return nil
}
