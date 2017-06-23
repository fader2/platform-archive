package objects

import (
	"io"

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
	Meta() Meta
	SetMeta(Meta)
	Size() int64
	SetSize(int64)
	Reader() (io.ReadCloser, error)
	Writer() (io.WriteCloser, error)
}
