package objects

import uuid "github.com/satori/go.uuid"

func UUIDFromString(name string) uuid.UUID {
	return uuid.NewV5(uuid.NamespaceOID, name)
}
