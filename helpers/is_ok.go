package helpers

import "reflect"

func IsOk[O any](val O) (O, bool) {
	var zero O
	return val, !reflect.DeepEqual(val, zero)
}
