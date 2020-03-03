# protostructure [![Godoc](https://godoc.org/github.com/mitchellh/protostructure?status.svg)](https://godoc.org/github.com/mitchellh/protostructure)

protostructure is a Go library for encoding and decoding a `struct`
_type_ over the wire.

This library is useful when you want to send arbitrary structures
over protocol buffers for behavior such as configuration decoding
(`encoding/json`, etc.), validation (using packages that use tags), etc.
This works because we can reconstruct the struct type dynamically using
`reflect` including any field tags.

This library only sends the structure of the struct, not the _value_.
If you want to send the value, you should build your protocol buffer
message in such a way that it encodes that somehow using something
such as JSON.

## Installation

Standard `go get`:

```
$ go get github.com/mitchellh/protostructure
```

## Usage & Example

For usage and examples see the [Godoc](http://godoc.org/github.com/mitchellh/protostructure).

A quick code example is shown below using both the imaginary proto file
and the Go code that uses it.

```proto
syntax = "proto3";
package myapp;
import "protostructure.proto";

// Response is an example response structure for an RPC endpoint.
message Response {
	protostructure.Struct config = 1;
}
```

```go
type Config struct {
	Name string            `json:"name"`
	Meta map[string]string `json:"metadata"`
	Port []*Port           `json:"ports"`
}

type Port struct {
	Number uint `json:"number"`
	Desc   string `json:"desc"`
}

// You can encode the structure on one side:
message, err := protostructure.Encode(Config{})

// And you can use the structure on the other side. Imagine resp
// is populated using some protobuf RPC such as gRPC.
val, err := protostructure.New(resp.Config)
json.Unmarshal([]byte(`{
	"name": "example",
	"meta": { "env": "prod" },
	"ports": [
		{ "number": 8080 },
		{ "number": 8100, desc: "backup" },
	]
}`, val)

// val now holds the same structure dynamically. You can pair with other
// libraries such as https://github.com/go-playground/validator to also
// send validation using this library.
```

## Limitations

There are several limitations on the structures that can be encoded.
Some of these limitations are fixable but the effort hasn't been put in
while others are fundamental due to the limitations of Go currently:

  * Circular references are not allowed between any struct types.
  * Embedded structs are not supported
  * Methods are not preserved, and therefore interface implementation
    is not known. This is also an important detail because custom callbacks
    such as `UnmarshalJSON` may not work properly.
  * Field types cannot be: interfaces, channels, functions
  * Certain stdlib types such as `time.Time` currently do not encode well.

## But... why?

The real world use case that led to the creation of this library was
to facilitate decoding and validating configuration for plugins via
[go-plugin](https://github.com/hashicorp/go-plugin), a plugin system for
Go that communicates using [gRPC](https://grpc.io).

The plugins for this particular program have dynamic configuration structures
that were decoded using an `encoding/json`-like interface (struct tags) and
validated using [`go-playground/validator`](https://github.com/go-playground/validator)
which also uses struct tags. Using protostructure, we can send the configuration
structure across the wire, decode and validate the configuration in the host process,
and report more rich errors that way.

Another reason we wanted to ship the config structure vs. ship the
config is because the actual language we are using for configuration
is [HCL](https://github.com/hashicorp/hcl) which supports things like
function calls, logic, and more and shipping that runtime across is
much, much more difficult.

This was extracted into a separate library because the ability to
encode a Go structure (particulary to include tags) seemed more generally
useful, although rare.
