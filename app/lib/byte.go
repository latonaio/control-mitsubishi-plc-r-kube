package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"unsafe"
)

func Byte2String(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func Byte2Int(arr []byte) int {
	val := int64(0)
	size := len(arr)
	for i := 0; i < size; i++ {
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

func GetBytesFrom32bitWithLE(i int64) []byte {
	buf := new(bytes.Buffer)
	n, _ := strconv.Atoi(strconv.FormatInt(i, 16))

	err := binary.Write(buf, binary.LittleEndian, int32(n))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func GetBytesFrom8bitWithLE(i int64) []byte {
	buf := new(bytes.Buffer)
	n, _ := strconv.Atoi(strconv.FormatInt(i, 16))
	err := binary.Write(buf, binary.LittleEndian, int8(n))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}
