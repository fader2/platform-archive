package objects

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/fader2/platform/utils"
	lua "github.com/yuin/gopher-lua"
)

const (
	META_ORIGINAL_NAME = "Original-Name"
	META_CONTENT_TYPE  = "Content-Type"
)

func (b *Blob) SetOrigName(name string) {
	b.Meta.Set(META_ORIGINAL_NAME, name)
}

// SetIDFromName устанавливает ID объекта на основе его имени
func (b *Blob) SetIDFromName(
	name string,
) {
	b.ID = UUIDFromString(name)
}

// SetDataFromValue устанавливает данные объекта на основе типа данных
func (b *Blob) SetDataFromValue(
	in interface{},
) error {
	ct := TypeFrom(in)
	b.Meta.Set(META_CONTENT_TYPE, ct.String())

	switch ct {
	case TString:
		b.Data = []byte(in.(string))
	case TNumber:
		b.Data = utils.Float64bytes(utils.ToFloat64(in))
	case TBool:
		if in.(bool) {
			b.Data = utils.Float64bytes(1)
		} else {
			b.Data = utils.Float64bytes(0)
		}
	case TArray:
		var buf = new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(in.([]interface{})); err != nil {
			return errors.New("encode array: " + err.Error())
		}

		b.Data = buf.Bytes()
	case TMap:
		var buf = new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(in.(map[string]interface{})); err != nil {
			return errors.New("encode map: " + err.Error())
		}

		b.Data = buf.Bytes()
	default:
		return errors.New("not supported content type " + ct.String())
	}

	return nil
}

// PushDataToLState передать данные объекта в lua
func (b *Blob) PushDataToLState(L *lua.LState) int {
	ct := b.Meta.Get(META_CONTENT_TYPE)
	switch ContentType(ct) {
	case TNumber:
		v := utils.Float64frombytes(b.Data)
		L.Push(lua.LNumber(v))
		return 1
	case TString:
		L.Push(lua.LString(string(b.Data)))
		return 1
	case TBool:
		v := utils.Float64frombytes(b.Data)
		L.Push(lua.LBool(v == 1))
		return 1
	case TArray:
		var arr []interface{}
		dec := gob.NewDecoder(bytes.NewBuffer(b.Data))
		if err := dec.Decode(&arr); err != nil {
			L.RaiseError("decode array %s", err)
			return 0
		}
		lv := utils.ToLValueOrNil(arr, L)
		L.Push(lv)
		return 1
	case TMap:
		var m map[string]interface{}
		dec := gob.NewDecoder(bytes.NewBuffer(b.Data))
		if err := dec.Decode(&m); err != nil {
			L.RaiseError("decode array %s", err)
			return 0
		}
		lv := utils.ToLValueOrNil(m, L)
		L.Push(lv)
		return 1
	default:
		L.RaiseError("not supported content type %q", ct)
	}

	return 0
}
