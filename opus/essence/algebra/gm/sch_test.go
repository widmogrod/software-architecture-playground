package gm

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
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
	Id      *string `name:"id"`
	Content *string `name:"content"`
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
	if v.Kind() == reflect.Struct {
		msg := strings.Builder{}
		foundFields := make(map[string]bool)
		// iterate over fields
		for i := 0; i < v.NumField(); i++ {
			// check if field is required
			field := v.Type().Field(i)
			// get field from schema
			fieldName := field.Name
			attr, ok := sch.Attrs[fieldName]
			if !ok {
				// field name from tag name
				fieldName = field.Tag.Get("name")
				attr, ok = sch.Attrs[fieldName]
				if !ok {
					msg.WriteString("field " + fieldName + " is not defined in schema; ")
					continue
				}
			}

			foundFields[fieldName] = true

			// check if field is required
			if attr.Required {
				// check if field is set
				if v.Field(i).IsZero() {
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

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]*Schema),
	}
}
