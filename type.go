package protostructure

import (
	"fmt"
	"reflect"
)

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

// goType returns the reflect.Type for a *Type value.
func goType(t *Type) reflect.Type {
	switch t := t.Type.(type) {
	case *Type_Primitive:
		// Look up the type directly by kind
		return kindTypes[reflect.Kind(t.Primitive.Kind)]

	case *Type_Struct:
		// Build the type
		return reflectType(t.Struct)

	case *Type_Container:
		switch reflect.Kind(t.Container.Kind) {
		case reflect.Map:
			return reflect.MapOf(goType(t.Container.Key), goType(t.Container.Elem))

		case reflect.Ptr:
			return reflect.PtrTo(goType(t.Container.Elem))

		case reflect.Slice:
			return reflect.SliceOf(goType(t.Container.Elem))

		case reflect.Array:
			return reflect.ArrayOf(int(t.Container.Count), goType(t.Container.Elem))
		}
	}

	panic(fmt.Sprintf("unknown type to decode: %#v", t))
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

// kindTypes is a mapping from reflect.Kind to an equivalent reflect.Type
// representing that kind. This only contains the "primitive" values as
// defined by what a valid value is for a Type_Primitive.
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
	reflect.Interface:  reflect.TypeOf((*interface{})(nil)).Elem(),
	reflect.String:     reflect.TypeOf(""),
}
