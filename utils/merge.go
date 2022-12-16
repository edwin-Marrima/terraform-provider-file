package utils

import (
	"fmt"
	"reflect"
)

type Mergito struct {
	Src           any
	Dst           any
	OverrideArray bool
}

func WithOverrideArray(append bool) func(*Mergito) {
	return func(m *Mergito) {
		m.OverrideArray = append
	}
}

func Merge(src any, dst any, options ...func(*Mergito)) (any, error) {
	m := &Mergito{Src: src, Dst: dst, OverrideArray: false}
	for _, opt := range options {
		opt(m)
	}
	a, err := DeepMerge(reflect.ValueOf(m.Src), reflect.ValueOf(m.Dst), m.OverrideArray)
	return a, err
}

func DeepMerge(src, dst reflect.Value, overrideArray bool) (any, error) {
	if src.Kind() != reflect.Map || dst.Kind() != reflect.Map {
		return dst.Interface(), nil
	}

	// iterate over src map, these interactions aim to get the src map keys in
	// order to verify if they exist in the destination map
	iter := src.MapRange()
	for iter.Next() {

		// Map key and value
		srcMapKey := iter.Key()
		srcMapValue := reflect.ValueOf(iter.Value().Interface())
		// extract the value associated with the key in the destination map. If the key does not exist
		// dstMapValue.Kind() will return reflect.Invalid type
		dstMapValue := dst.MapIndex(srcMapKey)

		switch {
		case srcMapValue.Kind() == reflect.Map && dstMapValue.Kind() != reflect.Invalid:
			if s := dataTypeValidation(srcMapValue.Type(), dstMapValue.Elem().Type()); s != "" {
				return nil, fmt.Errorf(s)
			}
			DeepMerge(srcMapValue, dstMapValue.Elem(), overrideArray)
			continue
		case srcMapValue.Kind() == reflect.Slice && dstMapValue.Kind() != reflect.Invalid:
			if overrideArray {
				dst.SetMapIndex(srcMapKey, srcMapValue)
				continue
			}
			if s := dataTypeValidation(srcMapValue.Type(), dstMapValue.Elem().Type()); s != "" {
				return nil, fmt.Errorf(s)
			}
			dst.SetMapIndex(srcMapKey, reflect.AppendSlice(dstMapValue.Elem(), srcMapValue))
			continue
		}

		dst.SetMapIndex(srcMapKey, srcMapValue)
	}
	return dst.Interface(), nil
}

func dataTypeValidation(src, dst reflect.Type) string {
	if src != dst {
		return fmt.Sprintf("Cannot append two %ss with different type (%s, %s)", src.Kind(), src, dst)
	}
	return ""

}
