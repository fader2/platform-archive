package objects

import (
	"bytes"
	fmt "fmt"
	io "io"

	"github.com/fader2/platform/consts"
	uuid "github.com/satori/go.uuid"
)

var _ Storer = (*testStore)(nil)

func newTestStore() *testStore {
	return &testStore{
		store: make(map[uuid.UUID][]byte),
	}
}

type testStore struct {
	store map[uuid.UUID][]byte
}

func (s *testStore) NewEncodedObject(id uuid.UUID) EncodedObject {
	return NewObject(id)
}
func (s *testStore) EncodedObject(
	_type ObjectType,
	id uuid.UUID,
) (EncodedObject, error) {
	var buf = new(bytes.Buffer)
	buf.Write(s.store[id])

	if buf.Len() == 0 {
		return nil, consts.ErrNotFound
	}
	or := NewReader(buf)
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
func (s *testStore) SetEncodedObject(obj EncodedObject) (
	id uuid.UUID,
	err error,
) {
	id = obj.ID()
	buf := new(bytes.Buffer)
	ow := NewWriter(buf)
	if err := ow.WriteHeader(obj.Type(), obj.Meta()); err != nil {
		return id, fmt.Errorf("write header %s", err)
	}

	r, _ := obj.Reader()
	io.Copy(ow, r)

	s.store[id] = buf.Bytes()

	return
}
