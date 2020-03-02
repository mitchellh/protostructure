// Package protostructure provides a mechanism for encoding and decoding
// a struct _type_ using protocol buffers. To be clear: this encodes the
// _type_ and not the _value_.
//
// Most importantly, this lets you do things such as transferring a struct
// that supports JSON decoding across a protobuf RPC, and then decoding
// a JSON value directly into it since you have access to things such as
// struct tags from the remote end.
//
// For a pure JSON use case, it may make sense to instead send the JSON
// rather than send the struct type. There are other scenarios where sending
// the type is easier and this library facilitates those use cases.
//
// The primary functions you want to look at are "Encode" and "New".
package protostructure

import (
	"fmt"
	"reflect"
)

//go:generate sh -c "protoc ./*.proto --go_out=plugins=grpc:./"

// Encode converts a struct to a *Struct which implements proto.Message
// and can therefore be sent over the wire. Note that only the _structure_
// of the struct is encoded and NOT any fields values.
//
// Encoding has a number of limitations:
//
//   * Circular references are not allowed between any struct types
//   * Embedded structs are not supported
//   * Methods are not preserved
//   * Field types cannot be: interfaces, channels, functions
//
func Encode(s interface{}) (*Struct, error) {
	// If s is a Type already then we use that directly (a code path used
	// by protoType but not generally expected for callers).
	t, ok := s.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(s)
	}

	// First we need to unwrap any number of layers of interface{} or pointers.
	for {
		// If we don't have some container type, then we are done.
		if t.Kind() != reflect.Ptr && t.Kind() != reflect.Interface {
			break
		}

		// Unwrap one layer
		t = t.Elem()
	}

	// We require a struct since that's what we're encoding here.
	if k := t.Kind(); k != reflect.Struct {
		return nil, fmt.Errorf("encode: requires a struct, got %s", k)
	}

	// Build our struct
	var result Struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType, err := protoType(field.Type)
		if err != nil {
			return nil, err
		}

		result.Fields = append(result.Fields, &Struct_Field{
			Name:    field.Name,
			PkgPath: field.PkgPath,
			Tag:     string(field.Tag),
			Type:    fieldType,
		})
	}

	return &result, nil
}

// New returns a pointer to an allocated struct for the structure given
// or an error if there are any invalid fields.
//
// This interface{} value can be used directly in functions such as
// json.Unmarshal, or it can be inspected further as necessary.
func New(s *Struct) (result interface{}, err error) {
	// We need to use the recover mechanism because the primary source
	// of underlying errors is the stdlib reflect library which just panics
	// whenever there is invalid input.
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	return reflect.New(reflectType(s)).Interface(), nil
}
