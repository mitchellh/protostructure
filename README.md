# protostructure [![Godoc](https://godoc.org/github.com/mitchellh/protostructure?status.svg)](https://godoc.org/github.com/mitchellh/protostructure)

protostructure is a Go library for encoding and decoding a `struct`
_type_ over the wire.

This library is useful when you want to send arbitrary structures
over a protocol buffer RPC for behavior such as configuration decoding
(`encoding/json`, etc.), validation (using packages that use tags), etc.
This works because we can reconstruct the struct type dynamically using
`reflect` including any field tags.

## Installation

Standard `go get`:

```
$ go get github.com/mitchellh/protostructure
```

## Usage & Example

For usage and examples see the [Godoc](http://godoc.org/github.com/mitchellh/protostructure).
