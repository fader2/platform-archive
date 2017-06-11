package boltdb

import (
	"io"

	"bytes"

	"fmt"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"
)

type Storer interface {
	NewEncodedObject(uuid.UUID) EncodedObject
	EncodedObject(ObjectType, uuid.UUID) (EncodedObject, error)
	SetEncodedObject(obj EncodedObject) (uuid.UUID, error)
}

type EncodedObject interface {
	ID() uuid.UUID
	Type() ObjectType
	SetType(ObjectType)
	ContentType() string
	SetContentType(string)
	Size() int64
	SetSize(int64)
	Reader() (io.ReadCloser, error)
	Writer() (io.WriteCloser, error)
}

var _ Storer = (*BoltdbStorage)(nil)

func NewBlobStorage(db *bolt.DB, name string) *BoltdbStorage {
	return &BoltdbStorage{db, name}
}

type BoltdbStorage struct {
	db     *bolt.DB
	bucket string
}

func (o *BoltdbStorage) NewEncodedObject(id uuid.UUID) EncodedObject {
	return &obj{
		id: id,
	}
}

func (s *BoltdbStorage) EncodedObject(_type ObjectType, id uuid.UUID) (EncodedObject, error) {
	var buf = new(bytes.Buffer)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		_, err := buf.Write(b.Get(id.Bytes()))
		return err
	})
	if err != nil {
		return nil, err
	}
	or := newReader(buf)
	defer or.Close()
	gotType, gotContentType, err := or.Header()
	if err != nil {
		return nil, fmt.Errorf("read header %s", err)
	}

	if gotType != _type {
		return nil, fmt.Errorf("not expected obj type, got type %s", gotType)
	}

	obj := s.NewEncodedObject(id)
	obj.SetType(gotType)
	obj.SetSize(int64(buf.Len()))
	obj.SetContentType(gotContentType)
	w, err := obj.Writer()
	if err != nil {
		return nil, fmt.Errorf("get object writer %s", err)
	}
	defer w.Close()

	io.Copy(w, or)

	return obj, nil
}

func (s *BoltdbStorage) SetEncodedObject(obj EncodedObject) (
	id uuid.UUID,
	err error,
) {
	id = obj.ID()
	buf := new(bytes.Buffer)
	ow := newWriter(buf)
	if err := ow.WriteHeader(obj.Type(), obj.ContentType()); err != nil {
		return id, fmt.Errorf("write header %s", err)
	}

	r, _ := obj.Reader()
	io.Copy(ow, r)

	err = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(s.bucket))
		if err != nil {
			return err
		}
		return b.Put(id.Bytes(), buf.Bytes())
	})

	return
}
