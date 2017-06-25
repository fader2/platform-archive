package boltdb

import (
	"io"

	"bytes"

	"fmt"

	"github.com/boltdb/bolt"
	"github.com/fader2/platform/consts"
	"github.com/fader2/platform/objects"
	uuid "github.com/satori/go.uuid"
)

var _ objects.Storer = (*BoltdbStorage)(nil)

func NewBlobStorage(db *bolt.DB, name string) (s *BoltdbStorage) {
	s = &BoltdbStorage{db, name}
	return
}

type BoltdbStorage struct {
	db     *bolt.DB
	bucket string
}

func (o *BoltdbStorage) NewEncodedObject(id uuid.UUID) objects.EncodedObject {
	return objects.NewObject(id)
}

func (s *BoltdbStorage) EncodedObject(_type objects.ObjectType, id uuid.UUID) (objects.EncodedObject, error) {
	var buf = new(bytes.Buffer)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		_, err := buf.Write(b.Get(id.Bytes()))
		return err
	})
	if err != nil {
		return nil, err
	}
	if buf.Len() == 0 {
		return nil, consts.ErrNotFound
	}
	or := objects.NewReader(buf)
	defer or.Close()
	gotType, gotMeta, err := or.Header()
	if err != nil {
		return nil, fmt.Errorf("read header %s", err)
	}

	if gotType != _type {
		return nil, fmt.Errorf("not expected obj type, got type %s", gotType)
	}

	obj := s.NewEncodedObject(id)
	obj.SetType(gotType)
	obj.SetMeta(gotMeta)
	obj.SetSize(int64(buf.Len()))
	w, err := obj.Writer()
	if err != nil {
		return nil, fmt.Errorf("get object writer %s", err)
	}
	defer w.Close()

	io.Copy(w, or)

	return obj, nil
}

func (s *BoltdbStorage) SetEncodedObject(obj objects.EncodedObject) (
	id uuid.UUID,
	err error,
) {
	id = obj.ID()
	buf := new(bytes.Buffer)
	ow := objects.NewWriter(buf)
	if err := ow.WriteHeader(obj.Type(), obj.Meta()); err != nil {
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
