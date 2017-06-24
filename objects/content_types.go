package objects

import (
	"strings"
)

const (
	PREFIX_FADERDATATYPE = "application/fader.datatypes."
)

type ContentType string

var (
	TUnknown ContentType = "application/fader.dt.unknown"
	TString  ContentType = "application/fader.dt.string"
	TNumber  ContentType = "application/fader.dt.number"
	TBool    ContentType = "application/fader.dt.bool"
	TArray   ContentType = "application/fader.dt.array"
	TMap     ContentType = "application/fader.dt.map"
	TCustom  ContentType = "application/fader.dt.custom."
)

func TypeFrom(v interface{}) (t ContentType) {
	switch v.(type) {
	case float32, float64, int, int32, int64:
		return TNumber
	case string:
		return TString
	case bool:
		return TBool
	case []interface{}:
		return TArray
	case map[string]interface{}:
		return TMap
	}

	return TUnknown
}

//

func (t ContentType) String() string {
	return string(t)
}

func (t ContentType) Valid() bool {
	return strings.HasPrefix(t.String(), PREFIX_FADERDATATYPE)
}

func (t ContentType) IsString() bool {
	if t.Valid() {
		return false
	}
	return strings.HasSuffix(t.String(), "string")
}

func (t ContentType) IsNumber() bool {
	if t.Valid() {
		return false
	}
	return strings.HasSuffix(t.String(), "number")
}

func (t ContentType) IsBool() bool {
	if t.Valid() {
		return false
	}
	return strings.HasSuffix(t.String(), "bool")
}

func (t ContentType) IsCustom() bool {
	if t.Valid() {
		return false
	}
	if len(t.String()) <= len(PREFIX_FADERDATATYPE+".custom.") {
		return false
	}
	return strings.HasPrefix(t.String(), PREFIX_FADERDATATYPE+".custom.")
}
