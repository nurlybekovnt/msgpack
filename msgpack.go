package msgpack

import "fmt"

type unexpectedCodeError struct {
	code byte
	hint string
}

func (err unexpectedCodeError) Error() string {
	return fmt.Sprintf("msgpack: unexpected code=%x decoding %s", err.code, err.hint)
}
