package string

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func ToString[T any](obj T) string {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Sprint(obj)
	}

	typ := val.Type()
	var parts []string
	for i := 0; i < val.NumField(); i++ {
		name := typ.Field(i).Name
		value := val.Field(i).Interface()
		parts = append(parts, fmt.Sprintf("%s=%v", name, value))
	}

	return fmt.Sprintf("%s{%s}", typ.Name(), strings.Join(parts, ", "))
}

func ToJSON(v any) string {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(b)
}

func Contains() {

}
