package util

import (
	"reflect"
)

func OneofFieldRef(oneof interface{}) interface{} {
	val := reflect.ValueOf(oneof)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct || val.NumField() != 1 {
		return nil
	}

	fv := val.Field(0)
	if !fv.CanInterface() {
		return nil
	}

	return fv.Interface()
}
