package boltdb

import (
	"os"
	"testing"
	"time"

	"io/ioutil"

	"bytes"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"
)

func TestGetSetObject(t *testing.T) {
	defer func() {
		os.RemoveAll("_fortesting.db")
	}()

	db, err := bolt.Open(
		"_fortesting.db",
		0600,
		&bolt.Options{
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		t.Fatal("setup db", err)
	}

	// new object

	s := NewBlobStorage(db, "_______fortesting")
	obj := s.NewEncodedObject(uuid.NewV4())
	obj.SetType(BlobObject)
	w, _ := obj.Writer()
	wantData := []byte("abcd1234")
	w.Write(wantData)

	_, err = s.SetEncodedObject(obj)
	if err != nil {
		t.Fatal("add new object", err)
	}

	// get new object

	gotObj, err := s.EncodedObject(BlobObject, obj.ID())
	if err != nil {
		t.Fatal("got object by ID", err)
	}

	r, _ := gotObj.Reader()
	gotData, _ := ioutil.ReadAll(r)
	if !bytes.Equal(
		gotData,
		wantData,
	) {
		t.Fatal("not expected data", err)
	}
}
