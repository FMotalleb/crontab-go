package utils

import (
	"errors"
	"reflect"
)

func FirstNonZeroForced[T comparable](items ...T) T {
	item, _ := FirstNonZero(items...)
	return item
}

func FirstNonZero[T comparable](items ...T) (T, error) {
	z := reflect.Zero(reflect.TypeFor[T]()).Interface().(T)
	item := FirstThatMatch(
		func(i T) bool {
			return i != z
		},
		z,
		items...,
	)
	if item == z {
		return z, errors.New("no non-zero element where found in given items")
	}
	return item, nil
}

func FirstThatMatch[T comparable](test func(item T) bool, defaultValue T, items ...T) T {
	for _, item := range items {
		if test(item) {
			return item
		}
	}
	return defaultValue
}
