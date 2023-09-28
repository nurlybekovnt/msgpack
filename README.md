# Efficient MessagePack encoding for Golang

## Overview

This project is an implementation of the MessagePack (Msgpack) encoding format
in the Go programming language. It is inspired by the
`github.com/vmihailenco/msgpack` package and is optimized for zero memory
allocations and maximum speed.

## Key features

- **High Performance with Zero Allocations**
- **Append-Style API:** The package follows an append-style API pattern similar
  to `hex.AppendEncode`, `strconv.AppendBool`, and `fmt.Appendf` from the
  standard library. For example:
    - `func (e *Encoder) AppendInt8(dst []byte, n int8) []byte`
    - `func (e *Encoder) AppendBytes(dst []byte, v []byte) []byte {`
    - `func (e *Encoder) AppendString(dst []byte, v string) []byte {`
    - `func (e *Encoder) AppendMap(dst []byte, m map[string]interface{}) []byte {`

## Usage

To use this package in your Go project, follow these steps:

1. Install msgpack

```bash
go get github.com/nurlybekovnt/msgpack
```


## Acknowledgment

This project builds upon the work of `github.com/vmihailenco/msgpack`. We extend
my gratitude to the creators and contributors of that project, which has been
instrumental in the development of this implementation.

## Contributions

We welcome contributions and bug reports. If you have any suggestions or
improvements, please feel free to open an issue or submit a pull request on our
GitHub repository.
