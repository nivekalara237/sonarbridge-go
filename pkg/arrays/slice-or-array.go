package arrays

import (
	"cmp"
	"reflect"
	"slices"
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

// GroupBy regroupe les éléments de slice selon la clé retournée par keyFn.
// L'ordre des clés n'est pas garanti (comportement map standard).
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, item := range slice {
		k := keyFn(item)
		out[k] = append(out[k], item)
	}
	return out
}

// GroupByOrdered regroupe et retourne les clés triées.
// Renvoie un slice d'entrées {Key, Items} dans l'ordre croissant des clés.
func GroupByOrdered[T any, K cmp.Ordered](slice []T, keyFn func(T) K) []Group[K, T] {
	m := make(map[K][]T)
	for _, item := range slice {
		k := keyFn(item)
		m[k] = append(m[k], item)
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	out := make([]Group[K, T], 0, len(keys))
	for _, k := range keys {
		out = append(out, Group[K, T]{Key: k, Items: m[k]})
	}
	return out
}

type Group[K comparable, T any] struct {
	Key   K
	Items []T
}
