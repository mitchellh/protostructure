package protostructure

import (
	"fmt"
	"reflect"
)

//go:generate sh -c "protoc ./*.proto --go_out=plugins=grpc:./"

// Encode converts a struct to a *Struct which implements proto.Message
// and can therefore be sent over the wire. Note that only the _structure_
// of the struct is encoded and NOT any fields values.
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
