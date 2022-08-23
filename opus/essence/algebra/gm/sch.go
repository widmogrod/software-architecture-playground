package gm

import (
	"errors"
	"reflect"
	"strings"
	"sync"
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

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]*Schema),
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
