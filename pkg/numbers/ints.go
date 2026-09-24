package numbers

import (
	"fmt"
	"strconv"
)

func ToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func ToInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case string:
		{
			i64, err := strconv.ParseInt(v, 10, 32)
			return int32(i64), err
		}
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}
