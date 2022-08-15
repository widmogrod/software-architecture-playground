package lets_build_db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type (
	Schema struct {
		Fields []Field
	}
	Field struct {
		Name string
		Typ  TypeOf
	}
	TypeOf struct {
		Identifier *Identifier
		String     *String
		DateTime   *DateTime
	}

	Identifier struct{}
	String     struct{}

	DateTime struct {
		Format     string
		DefaultNow bool
	}
)

func NewDB(schema Schema) *DB {
	i := 0
	return &DB{
		schema:    schema,
		appendLog: nil,
		nextRowId: func() string {
			i++
			return strconv.Itoa(i)
		},
	}
}

type (
	DB struct {
		schema    Schema
		appendLog AppendLog
		nextRowId func() string
	}

	Predicate struct {
		Eq  *LeftOnly
		And *LeftRight
		Or  *LeftRight
	}
	LeftRight struct {
		L Predicate
		R Predicate
	}
	LeftOnly struct {
		Field string
		Value interface{}
	}
)

func (d *DB) InsertInto(m map[string]interface{}) (error, string) {
	if ok, problem := d.Validate(m); !ok {
		return errors.New(problem), ""
	}

	rowId := d.nextRowId()
	d.appendLog = Set(d.appendLog, d.toSortedSet(m, rowId))
	return nil, rowId
}

func (d *DB) toSortedSet(m map[string]interface{}, rowId string) KVSortedSet {
	var result KVSortedSet

	fields := d.schemaFieldsAsMap()
	for fieldName, fieldValue := range m {
		result = append(result, KV{
			fmt.Sprintf("%s:%s", fieldName, rowId),
			d.serialiseValue(fieldValue, fields[fieldName]),
		})
	}

	return result
}

func (d *DB) serialiseValue(value interface{}, field Field) string {
	//if field.Typ.Identifier != nil {
	//	// Identifier can be only set by DB, and type assumption is safe
	//	return strconv.Itoa(value.(int))
	//}

	if field.Typ.String != nil {
		// Serialisation of value must pass validation function
		return value.(string)
	}

	if field.Typ.DateTime != nil {
		return value.(time.Time).Format(field.Typ.DateTime.Format)
	}

	panic("serialisation failure, should never happen!")
}

func (d *DB) Validate(m map[string]interface{}) (bool, string) {
	fields := d.schemaFieldsAsMap()

	for fieldName, fieldValue := range m {
		if field, ok := fields[fieldName]; ok {
			valid, problem := d.validateField(field, fieldValue)
			if !valid {
				return valid, problem
			}
		} else {
			return false, fmt.Sprintf(`field "%s" does not exist in schema`, fieldName)
		}
	}

	return true, ""
}

func (d *DB) validateField(field Field, fieldValue interface{}) (bool, string) {
	if field.Typ.Identifier != nil {
		return false, fmt.Sprintf(`field "%s" is identified and cannot be set!`, field.Name)
	} else if field.Typ.String != nil {
		if _, ok := fieldValue.(string); !ok {
			return false, fmt.Sprintf(`field "%s" is expected to have string value but "%#v" given`, field.Name, fieldValue)
		}
	} else if field.Typ.DateTime != nil {
		if _, ok := fieldValue.(time.Time); !ok {
			return false, fmt.Sprintf(`field "%s" is expected to be "*time.Time" value but "%#v" given`, field.Name, fieldValue)
		}
	}

	return true, ""
}

func (d *DB) schemaFieldsAsMap() map[string]Field {
	fields := make(map[string]Field)
	for _, field := range d.schema.Fields {
		fields[field.Name] = field
	}
	return fields
}

type Cursor struct {
}

func (d *DB) Select(p Predicate) map[string]interface{} {
	fields := d.schemaFieldsAsMap()
	kvSet := Find(d.appendLog, func(kv KV) bool {
		composeKey := kv[KEY]
		key := strings.Split(composeKey, ":")[0]
		val := kv[VAL]

		if p.Eq != nil {
			field, ok := fields[p.Eq.Field]
			if !ok {
				panic("field in predicate does not exits")
			}
			cand := d.serialiseValue(p.Eq.Value, field)
			if p.Eq.Field == key {
				return cand == val
			}
		}

		// TODO implement more!
		return false
	}, 4)

	result := make(map[string]interface{})
	for _, kv := range kvSet {
		composeKey := kv[KEY]
		key := strings.Split(composeKey, ":")[0]
		val := kv[VAL]

		result[key] = val
	}

	return result
}
