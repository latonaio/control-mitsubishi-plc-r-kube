package lib

import (
	"encoding/binary"
	"unsafe"
)

func Byte2String(bytes []byte) string{
	return *(*string)(unsafe.Pointer(&bytes))
}

func Byte2Int(arr []byte) int{
	val := int64(0)
	size := len(arr)
	for i := 0 ; i < size ; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return int(val)
}

func Int2bytes(i int, size int) []byte {
	var ui uint64
	if 0 < i {
		ui = uint64(i)
	} else {
		ui = (^uint64(-i) + 1)
	}
	return Uint2bytes(ui, size)
}

func Uint2bytes(i uint64, size int) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, i)
	return bytes[8-size : 8]
}
