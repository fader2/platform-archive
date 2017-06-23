package utils

import (
	"encoding/binary"
	"math"
)

func ToFloat64(v interface{}) (f float64) {
	switch _v := v.(type) {
	case int:
		f = float64(_v)
	case int16:
		f = float64(_v)
	case int32:
		f = float64(_v)
	case int64:
		f = float64(_v)
	case int8:
		f = float64(_v)
	case float32:
		f = float64(_v)
	case float64:
		f = float64(_v)
	case uint:
		f = float64(_v)
	case uint16:
		f = float64(_v)
	case uint32:
		f = float64(_v)
	case uint64:
		f = float64(_v)
	case uint8:
		f = float64(_v)
	default:
		f = 0
	}

	return
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
