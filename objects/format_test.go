package objects

import "testing"
import "bytes"
import "io/ioutil"

func TestReadWriterFormat(t *testing.T) {
	obj := new(bytes.Buffer)
	w := NewWriter(obj)
	meta := Meta{Meta: make(map[string]string)}
	meta.Set(META_CONTENT_TYPE, "text/plain")
	w.WriteHeader(BlobObject, meta)
	objData := []byte("abcdabcdabcdabcdabcd")
	w.Write(objData)

	newwork := obj.Bytes()

	got := bytes.NewBuffer(newwork)
	r := NewReader(got)
	gotType, gotMeta, err := r.Header()
	if err != nil {
		t.Fatal("read header", err)
	}
	if BlobObject != gotType {
		t.Fatal("not expected type", gotType)
	}
	if gotMeta.Get(META_CONTENT_TYPE) != "text/plain" {
		t.Fatal("not expected content type", gotMeta)
	}
	gotData, err := ioutil.ReadAll(got)
	if err != nil {
		t.Fatal("read data", err)
	}
	if !bytes.Equal(
		gotData,
		objData,
	) && len(gotData) > 0 {
		t.Fatal("not expected data", string(gotData))
	}
}
