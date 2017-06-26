package boltdb

import (
	"testing"

	"io/ioutil"

	"bytes"

	"github.com/fader2/platform/objects"
	uuid "github.com/satori/go.uuid"
)

func TestGetSetAnyObject(t *testing.T) {
	// new object

	s := NewBlobStorage(db, "_______fortesting")
	obj := s.NewEncodedObject(uuid.NewV4())
	obj.SetType(objects.BlobObject)
	w, _ := obj.Writer()
	wantData := []byte("abcd1234")
	w.Write(wantData)

	_, err := s.SetEncodedObject(obj)
	if err != nil {
		t.Fatal("add new object", err)
	}

	// get new object

	gotObj, err := s.EncodedObject(objects.BlobObject, obj.ID())
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

func TestBlob(t *testing.T) {

	blob := objects.EmptyBlob()
	blob.ID = uuid.NewV4()
	blob.Meta.Set(objects.META_CONTENT_TYPE, "text/plain")
	blob.Data = []byte("Abc")

	s := NewBlobStorage(db, "_______fortesting")

	_, err := objects.SetBlob(s, blob)
	if err != nil {
		t.Fatal("save blob by id", err)
	}
	wantData := blob.Data

	//

	blob, err = objects.GetBlob(s, blob.ID)
	if err != nil {
		t.Fatal("find blob by id", err)
	}
	if blob.Meta.Get(objects.META_CONTENT_TYPE) != "text/plain" {
		t.Fatal("not expected content type")
	}
	if len(blob.Data) == 0 {
		t.Fatal("empty data")
	}
	if !bytes.Equal(
		blob.Data,
		wantData,
	) {
		t.Fatal("not expected data")
	}
}

func TestUser(t *testing.T) {
	s := NewBlobStorage(db, "_______fortesting")

	u := objects.EmptyUser(objects.Client)
	u.ID = uuid.NewV4()
	u.Info.Pasport.Email = "client@test.com"
	u.Info.Profile.FirstName = "fname"

	id, err := objects.SetUser(s, u)
	if err != nil {
		t.Error("set user", err)
	}
	if id == uuid.Nil {
		t.Error("empty user ID")
	}

	//

	got, err := objects.GetUser(s, id)
	if err != nil {
		t.Error("find user", err)
	}

	if !objects.Client.Equal(objects.UserType(got.Meta.Get(objects.META_USER_TYPE))) {
		t.Error("not expected user type", got.Meta)
	}

	if got.Info.Pasport.Email != u.Info.Pasport.Email {
		t.Error("not expected email")
	}

	if got.Info.Profile.FirstName != u.Info.Profile.FirstName {
		t.Error("not expected first name")
	}
}
