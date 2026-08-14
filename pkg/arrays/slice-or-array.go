package arrays

import (
	"reflect"
)

func IsArrayOrSlice(value any) bool {
	return IsArray(value) || IsSlice(value)
}

func IsSlice(value any) bool {
	return reflect.ValueOf(value).Kind() == reflect.Slice
}

func IsArray(value any) bool {
	return reflect.ValueOf(value).Kind() == reflect.Array
}

func AnyToSlice(value any) {
	// return reflect.ValueOf(value).
}

func ContainsAnyOf[T any](arr []T, item T) bool {
	for _, el := range arr {
		if reflect.DeepEqual(el, item) {
			return true
		}
	}
	return false
}

func ContainsAnyOfFunc[T any](arr []T, item T, equal func(a, b T) bool) bool {
	for _, el := range arr {
		if equal(el, item) {
			return true
		}
	}
	return false
}
