package objects

import "testing"
import "bytes"
import "io/ioutil"

func TestReadWriterFormat(t *testing.T) {
	obj := new(bytes.Buffer)
	w := newWriter(obj)
	w.WriteHeader(BlobObject, "text/plain")
	objData := []byte("abcdabcdabcdabcdabcd")
	w.Write(objData)

	newwork := obj.Bytes()

	got := bytes.NewBuffer(newwork)
	r := newReader(got)
	gotType, gotContentType, err := r.Header()
	if err != nil {
		t.Fatal("read header", err)
	}
	if BlobObject != gotType {
		t.Fatal("not expected type", gotType)
	}
	if gotContentType != "text/plain" {
		t.Fatal("not expected content type", gotContentType)
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
