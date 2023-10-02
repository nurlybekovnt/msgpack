package msgpack

import "unsafe"

// bytesToString converts byte slice to a string without memory allocation. See
// https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
