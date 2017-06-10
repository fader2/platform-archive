package utils

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
