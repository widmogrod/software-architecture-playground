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
