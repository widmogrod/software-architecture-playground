package gm

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"testing"
)

func TestSchema(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register(QuestionSchema())
	assert.NoError(t, err)
	err = reg.Register(AnswerSchema())
	assert.NoError(t, err)
	err = reg.Register(CommentSchema())
	assert.NoError(t, err)

	// get schema
	sch, err := reg.Get("question")
	if assert.NoError(t, err) {
		assert.Equal(t, QuestionSchema(), *sch)
	}

	// validate object against schema
	err = reg.Validate("question", Question{})
	assert.ErrorContains(t, err, "field content is required")
	assert.ErrorContains(t, err, "field sourceId is required")
	assert.ErrorContains(t, err, "field sourceType is required")
	err = reg.Validate("question", Question{
		SourceId: kv.PtrString("unique-id"),
	})
	assert.ErrorContains(t, err, "field content is required")

}

func Sourcable(in map[string]AttrType) map[string]AttrType {
	// add sourceType and sourceId to the input map
	in["sourceType"] = AttrType{T: StringType, Required: true}
	in["sourceId"] = AttrType{T: StringType, Required: true, Identifier: true}
	return in
}

func Versionable(in map[string]AttrType) map[string]AttrType {
	// add version to the input map
	in["version"] = AttrType{T: IntType, Required: true, Default: 1}
	return in
}

func SchemaAware(schemaId string, in map[string]AttrType) map[string]AttrType {
	// add schema to the input map
	in["schema"] = AttrType{
		T:        StringType,
		Required: true,
		Default:  schemaId,
	}
	return in
}

func CommentSchema() Schema {
	return Schema{
		Name: "comment",
		Attrs: Sourcable(map[string]AttrType{
			"content": {T: StringType, Required: true},
		}),
	}
}

func AnswerSchema() Schema {
	return Schema{
		Name: "answer",
		Attrs: Sourcable(map[string]AttrType{
			"content": {T: StringType, Required: true},
		}),
	}
}

func QuestionSchema() Schema {
	schemaId := "question"
	return Schema{
		Name: schemaId,
		Attrs: SchemaAware(schemaId, Versionable(Sourcable(map[string]AttrType{
			"content": {T: StringType, Required: true},
		}))),
	}
}
