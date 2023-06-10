package gm

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"os"
	"testing"
)

func TestDSL(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register(QuestionSchema())
	assert.NoError(t, err)
	err = reg.Register(AnswerSchema())
	assert.NoError(t, err)
	err = reg.Register(CommentSchema())
	assert.NoError(t, err)

	// create file and write to it
	res := bytes.NewBuffer([]byte{})
	err = GenerateGolangCode(reg, "question", res, &GenConf{PackageName: "gm"})
	assert.NoError(t, err)

	err = os.WriteFile("dls_gen_question.go", res.Bytes(), 0644)
	assert.NoError(t, err)

	acl := NewGuard()
	err = acl.CreateRule("brainly.com", Predicate{
		Fields: map[string]Predicate{
			"action": {
				In: []interface{}{
					"CreateQuestionRequest",
				},
			},
		}})
	assert.NoError(t, err)

	store := kv.Default()

	dsl := NewStorageDSL(reg, acl, store, &Config{
		TenantId: "brainly.com",
	})
	err = dsl.Invoke(&CreateQuestionRequest{
		Id:      "1",
		Content: "How to write tests in golang?",
	})
	assert.NoError(t, err)
}
