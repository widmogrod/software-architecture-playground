package gm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type (
	Schema struct {
		Name  string
		Attrs map[string]AttrType
	}
)

type Question struct {
	Id         *string `name:"id"`
	Content    *string `name:"content"`
	Version    *int64  `name:"version"`
	SourceType *string `name:"sourceType"`
	SourceId   *string `name:"sourceId"`
}

func TestSchema(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register("question", QuestionSchema())
	assert.NoError(t, err)
	err = reg.Register("answer", AnswerSchema())
	assert.NoError(t, err)
	err = reg.Register("comment", CommentSchema())
	assert.NoError(t, err)

	// get schema
	sch, err := reg.Get("question")
	if assert.NoError(t, err) {
		assert.Equal(t, QuestionSchema(), *sch)
	}

	// validate object against schema
	err = reg.Validate("question", Question{})
	assert.ErrorContains(t, err, "field content is required")
	assert.ErrorContains(t, err, "field id is required")
	err = reg.Validate("question", Question{
		Id: kv.PtrString("unique-id"),
	})
	assert.ErrorContains(t, err, "field content is required")

}

func CommentSchema() Schema {
	return Schema{
		Name: "comment",
		Attrs: Sourcable(map[string]AttrType{
			"id":      {T: StringType, Required: true, Identifier: true},
			"content": {T: StringType, Required: true},
		}),
	}
}

func AnswerSchema() Schema {
	return Schema{
		Name: "answer",
		Attrs: Sourcable(map[string]AttrType{
			"id":      {T: StringType, Required: true, Identifier: true},
			"content": {T: StringType, Required: true},
		}),
	}
}

func QuestionSchema() Schema {
	return Schema{
		Name: "question",
		Attrs: Sourcable(map[string]AttrType{
			"id":      {T: StringType, Required: true, Identifier: true},
			"content": {T: StringType, Required: true},
		}),
	}
}

type SchemaRegistry struct {
	// schemas map
	schemas map[string]*Schema
	lock    sync.Mutex
}

func (r *SchemaRegistry) Register(id string, schema Schema) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	// check if id is unique and return error if not
	if _, ok := r.schemas[id]; ok {
		return errors.New("id " + id + " is not unique")
	}
	r.schemas[id] = &schema
	return nil
}

func (r *SchemaRegistry) Get(id string) (*Schema, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.schemas[id]; ok {
		return r.schemas[id], nil
	}
	return nil, errors.New("not implemented")
}

func (r *SchemaRegistry) Validate(id string, data interface{}) error {
	// get schema
	sch, err := r.Get(id)
	if err != nil {
		return err
	}

	// reflect data
	v := reflect.ValueOf(data)
	// if ptr to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		msg := strings.Builder{}
		foundFields := make(map[string]bool)
		// iterate over fields
		for i := 0; i < v.NumField(); i++ {
			// check if field is required
			field := v.Type().Field(i)
			attr, fieldName, err := extractAttr(field, sch)
			if err != nil {
				msg.WriteString(err.Error())
				continue
			}

			foundFields[*fieldName] = true

			// check if field is required
			if attr.Required {
				// check if field is set
				if v.Field(i).IsZero() {
					msg.WriteString("field " + *fieldName + " is required; ")
					continue
				}
			}
		}

		// check if all required fields are found
		for fieldName, attr := range sch.Attrs {
			if attr.Required && !foundFields[fieldName] {
				msg.WriteString("field " + fieldName + " is required; ")
			}
		}

		if msg.Len() > 0 {
			return errors.New(msg.String())
		}

		return nil
	}

	if v.Kind() == reflect.Map {
		msg := strings.Builder{}
		foundFields := make(map[string]bool)
		// iterate over fields
		for _, key := range v.MapKeys() {
			// check if field is required
			fieldName := key.String()
			// get attr from schema
			attr, ok := sch.Attrs[fieldName]
			if !ok {
				msg.WriteString("field " + fieldName + " is not defined in schema; ")
				continue
			}

			foundFields[fieldName] = true

			// check if field is required
			if attr.Required {
				// check if field is set
				if v.MapIndex(key).IsZero() {
					msg.WriteString("field " + fieldName + " is required; ")
					continue
				}
			}
		}

		// check if all required fields are found
		for fieldName, attr := range sch.Attrs {
			if attr.Required && !foundFields[fieldName] {
				msg.WriteString("field " + fieldName + " is required; ")
			}
		}

		if msg.Len() > 0 {
			return errors.New(msg.String())
		}

		return nil
	}

	// validate object against schema
	return errors.New("not implemented for this type" + v.Kind().String())
}

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

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]*Schema),
	}
}

func TestSchemaGeneration(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register("question", QuestionSchema())
	assert.NoError(t, err)

	obj := Question{}
	err = Generate(reg, "question", &obj)
	assert.NoError(t, err)

	err = reg.Validate("question", obj)
	assert.NoError(t, err)

	// json serialize obj
	b, err := json.Marshal(obj)
	assert.NoError(t, err)
	fmt.Println(string(b))

	obj2 := map[string]interface{}{}
	err = Generate(reg, "question", &obj2)
	assert.NoError(t, err)

	err = reg.Validate("question", obj2)
	assert.NoError(t, err)

	// json serialize obj
	b, err = json.Marshal(obj2)
	assert.NoError(t, err)
	fmt.Println(string(b))

}
