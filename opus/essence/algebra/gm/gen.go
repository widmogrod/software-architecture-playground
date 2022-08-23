package gm

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"math/rand"
	"reflect"
)

func Generate(r *SchemaRegistry, id string, in interface{}) error {
	// get schema
	sch, err := r.Get(id)
	if err != nil {
		return err
	}

	// reflect in
	v := reflect.ValueOf(in)
	// if ptr to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		// iterate over fields
		for i := 0; i < v.NumField(); i++ {
			// check if field is required
			field := v.Type().Field(i)
			attr, _, err := extractAttr(field, sch)
			if err != nil {
				return err
			}

			// randomly skip non required field
			if !attr.Required && rand.Intn(2) == 0 {
				continue
			}

			// field pointer
			fieldPtr := v.Field(i)

			// set field value depening of its type
			fieldPtr.Set(generateValueOfForType(attr))
		}

		return nil
	}

	if v.Kind() == reflect.Map {
		// iterate over attribute names
		for fieldName, attr := range sch.Attrs {
			// randomly skip non required field
			if !attr.Required && rand.Intn(2) == 0 {
				continue
			}
			v.SetMapIndex(reflect.ValueOf(fieldName), generateValueOfForType(&attr))
		}

		return nil
	}

	return errors.New("not implemented for this type: " + v.Kind().String())
}

func extractAttr(field reflect.StructField, sch *Schema) (*AttrType, *string, error) {
	// get field from schema
	fieldName := field.Name
	attr, ok := sch.Attrs[fieldName]
	if !ok {
		// field name from tag name
		fieldName = field.Tag.Get("name")
		attr, ok = sch.Attrs[fieldName]
		if !ok {
			return nil, nil, errors.New("field " + fieldName + " is not defined in schema")
		}
	}

	return &attr, &fieldName, nil
}

func generateValueOfForType(v *AttrType) reflect.Value {
	switch v.T {
	case StringType:
		return reflect.ValueOf(kv.PtrString(gofakeit.Word()))
	case IntType:
		return reflect.ValueOf(kv.PtrInt64(gofakeit.Int64()))
	case BoolType:
		return reflect.ValueOf(kv.PtrBool(gofakeit.Bool()))
	default:
		panic(fmt.Sprintf("not implemented for this type=%v", v))
	}
}
