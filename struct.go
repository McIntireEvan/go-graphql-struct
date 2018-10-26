package gqlstruct

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"time"
)

type GraphqlTyped interface {
	GraphqlType() graphql.Type
}

var GraphqlTypedType = reflect.TypeOf(new(GraphqlTyped)).Elem()

var timeType = reflect.TypeOf(time.Time{})

func fieldType(field reflect.StructField, v reflect.Value) graphql.Type {
	t := field.Type

	if t.Kind() == reflect.Struct {
		vStruct := v
		tStruct := t
		if vStruct.CanAddr() {
			vStruct = vStruct.Addr()
			tStruct = reflect.PtrTo(t)
		}
		if tStruct.Implements(GraphqlTypedType) {
			return vStruct.Interface().(GraphqlTyped).GraphqlType()
		}
	}

	if t.Implements(GraphqlTypedType) {
		return v.Interface().(GraphqlTyped).GraphqlType()
	}

	// Check if it is a pointer or interface...
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		// Updates the type with the type of the pointer
		t = t.Elem()
	}

	if t == timeType {
		return graphql.DateTime
	}

	switch t.Kind() {
	case reflect.Bool:
		return graphql.Boolean
	case reflect.String:
		return graphql.String
	case
		reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
		reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return graphql.Int
	case
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return graphql.Float
	}
	panic(fmt.Sprintf("%s not recognized", t))
}

func objectConfig(obj interface{}) graphql.ObjectConfig {
	fields := graphql.Fields{}

	val := reflect.ValueOf(obj).Elem()
	for i := 0; i < val.NumField(); i++ {
		fValue := val.Field(i)
		fType := val.Type().Field(i)
		tag, ok := fType.Tag.Lookup("graphql")
		if !ok {
			continue
		}

		t := fieldType(fType, fValue)
		if len(tag) > 0 && tag[0] == '!' {
			t = graphql.NewNonNull(t)
			tag = tag[1:]
		}
		fields[tag] = &graphql.Field{
			Type: t,
		}
	}

	return graphql.ObjectConfig{
		Name:   val.Type().Name(),
		Fields: fields,
	}
}

// Struct returns a `*graphql.Object` with the description extracted from the
// obj passed.
//
// This method extracts the information needed from the fields of the obj
// informed. All fields tagged with "graphql" are added.
//
// The "graphql" tag can be defined as:
//
// ```
// type T struct {
//     field string `graphql:"fieldname"`
// }
// ```
//
// * fieldname: The name of the field.
func Struct(obj interface{}) *graphql.Object {
	return graphql.NewObject(objectConfig(obj))
}
