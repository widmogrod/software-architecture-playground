package gm

import (
	"github.com/stretchr/testify/assert"
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
		SourceId: "unique-id",
	})
	assert.ErrorContains(t, err, "field content is required")

}

func Sourcable(in map[string]AttrType) map[string]AttrType {
	// add sourceType and sourceId to the input map
	in["sourceType"] = AttrType{T: StringType, Required: true}
	in["sourceId"] = AttrType{T: StringType, Required: true, Identifier: true}

	// add schema version to the input map
	//in["schema"] = AttrType{T: IntType, Required: true}

	// add version to the input map
	in["version"] = AttrType{T: IntType, Required: true}
	return in
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
			//"id":      {T: StringType, Required: true, Identifier: true},
			"content": {T: StringType, Required: true},
		}),
	}
}

func QuestionSchema() Schema {
	return Schema{
		Name: "question",
		Attrs: Sourcable(map[string]AttrType{
			//"id":      {T: StringType, Required: true, Identifier: true},
			"content": {T: StringType, Required: true},
		}),
	}
}
