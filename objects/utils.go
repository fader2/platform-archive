package objects

import uuid "github.com/satori/go.uuid"

var (
	NS, _ = uuid.FromString("ad1a7f74-0092-4a3f-838c-c277c3b0a7b8")
)

// UUIDFromString UUIDv5 from a string
func UUIDFromString(name string) uuid.UUID {
	return uuid.NewV5(NS, name)
}
