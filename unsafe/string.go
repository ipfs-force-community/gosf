// Package unsafe provides unsafe methods for converting bytes to string and reverse
package unsafe

import (
	"reflect"
	"unsafe"
)

// String converts bytes to string in unsafe way
func String(b []byte) string {
	sliceh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringh := &reflect.StringHeader{
		Data: sliceh.Data,
		Len:  sliceh.Len,
	}
	s := (*string)(unsafe.Pointer(stringh))
	return *s
}

// Bytes converts string to bytes in unsafe way
func Bytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sh := &reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	b := (*[]byte)(unsafe.Pointer(sh))
	return *b
}
