package utils

import "strconv"

func ToFloat64OrZero(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	default:
		return 0
	}
}

func ToFloat32OrZero(v any) float32 {
	switch n := v.(type) {
	case float64:
		return float32(n)
	case float32:
		return n
	case int:
		return float32(float64(n))
	case int64:
		return float32(n)
	case int32:
		return float32(n)
	case string:
		f, _ := strconv.ParseFloat(n, 32)
		return float32(f)
	default:
		return 0
	}
}
