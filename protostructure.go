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

// reflectType returns the reflect.Type for a Struct. This will panic if
// there are any invalid values in Struct. This behavior is not ideal but
// is inherited primarily by the underlying "reflect" standard library.
//
// It is recommended that you use a defer with a recover around this function
// in most cases since Struct is probably sent over the wire.
func reflectType(s *Struct) reflect.Type {
	var fields []reflect.StructField
	for _, field := range s.Fields {
		fields = append(fields, reflect.StructField{
			Name:    field.Name,
			PkgPath: field.PkgPath,
			Tag:     reflect.StructTag(field.Tag),
			Type:    goType(field.Type),
		})
	}

	return reflect.StructOf(fields)
}

func goType(t *Type) reflect.Type {
	switch t := t.Type.(type) {
	case *Type_Primitive:
		return kindTypes[reflect.Kind(t.Primitive.Kind)]

	case *Type_Struct:
		return reflectType(t.Struct)

	case *Type_Container:
		switch reflect.Kind(t.Container.Kind) {
		case reflect.Map:
			return reflect.MapOf(goType(t.Container.Key), goType(t.Container.Elem))

		case reflect.Ptr:
			return reflect.PtrTo(goType(t.Container.Elem))

			// TODO: others
		}
	}

	panic(fmt.Sprintf("unknown type to decode: %#v", t))
}

var kindTypes = map[reflect.Kind]reflect.Type{
	reflect.Bool:       reflect.TypeOf(false),
	reflect.Int:        reflect.TypeOf(int(0)),
	reflect.Int8:       reflect.TypeOf(int8(0)),
	reflect.Int16:      reflect.TypeOf(int16(0)),
	reflect.Int32:      reflect.TypeOf(int32(0)),
	reflect.Int64:      reflect.TypeOf(int64(0)),
	reflect.Uint:       reflect.TypeOf(uint(0)),
	reflect.Uint8:      reflect.TypeOf(uint8(0)),
	reflect.Uint16:     reflect.TypeOf(uint16(0)),
	reflect.Uint32:     reflect.TypeOf(uint32(0)),
	reflect.Uint64:     reflect.TypeOf(uint64(0)),
	reflect.Uintptr:    reflect.TypeOf(uintptr(0)),
	reflect.Float32:    reflect.TypeOf(float32(0)),
	reflect.Float64:    reflect.TypeOf(float64(0)),
	reflect.Complex64:  reflect.TypeOf(complex64(0)),
	reflect.Complex128: reflect.TypeOf(complex128(0)),
	reflect.Interface:  reflect.TypeOf(interface{}(nil)),
	reflect.String:     reflect.TypeOf(""),
}

// protoType takes a reflect.Type can returns the proto *Type value.
// This will return an error if the type given cannot be represented by
// protocol buffer messages.
func protoType(t reflect.Type) (*Type, error) {
	switch k := t.Kind(); k {
	// Primitives
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Interface,
		reflect.String:
		return &Type{
			Type: &Type_Primitive{
				Primitive: &Primitive{
					Kind: uint32(k),
				},
			},
		}, nil

	case reflect.Array:
		elem, err := protoType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &Type{
			Type: &Type_Container{
				Container: &Container{
					Kind:  uint32(k),
					Elem:  elem,
					Count: int32(t.Len()),
				},
			},
		}, nil

	case reflect.Map:
		key, err := protoType(t.Key())
		if err != nil {
			return nil, err
		}

		elem, err := protoType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &Type{
			Type: &Type_Container{
				Container: &Container{
					Kind: uint32(k),
					Elem: elem,
					Key:  key,
				},
			},
		}, nil

	case reflect.Ptr, reflect.Slice:
		elem, err := protoType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &Type{
			Type: &Type_Container{
				Container: &Container{
					Kind: uint32(k),
					Elem: elem,
				},
			},
		}, nil

	case reflect.Struct:
		elem, err := Encode(t)
		if err != nil {
			return nil, err
		}

		return &Type{
			Type: &Type_Struct{
				Struct: elem,
			},
		}, nil

	default:
		return nil, fmt.Errorf("encode: cannot encode type: %s (kind = %s)", t.String(), k)
	}
}
