package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(v interface{}) []byte {
	hasher := sha256.New()
	switch _v := v.(type) {
	case string:
		hasher.Write([]byte(_v))
	case []byte:
		hasher.Write(_v)
	}

	return hasher.Sum(nil)
}

func SHA1(v interface{}) []byte {
	hasher := sha1.New()
	switch _v := v.(type) {
	case string:
		hasher.Write([]byte(_v))
	case []byte:
		hasher.Write(_v)
	}

	return hasher.Sum(nil)
}

func SHA1String(v interface{}) string {
	return hex.EncodeToString(SHA1(v))
}

func SHA256String(v interface{}) string {
	return hex.EncodeToString(SHA256(v))
}
