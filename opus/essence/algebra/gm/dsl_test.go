package gm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDSL(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register("question", QuestionSchema())
	assert.NoError(t, err)
	err = reg.Register("answer", AnswerSchema())
	assert.NoError(t, err)
	err = reg.Register("comment", CommentSchema())
	assert.NoError(t, err)

	acl := NewGuard()
	err = acl.CreateRule("brainly.com", Predicate{Fields: map[string]Predicate{
		"action": {
			In: []interface{}{
				"CreateRequest",
				"UpdateRequest",
				"ArchiveRequest",
			},
		},
	}})
	assert.NoError(t, err)

	dsl := NewStorageDSL(reg, acl, &Config{
		tenantId: "brainly.com",
	})
	err = dsl.Invoke(&CreateRequest{
		Entity: "question",
		Data: map[string]interface{}{
			"sourceType": "brainly",
			"sourceId":   "1",
			"content":    "who was king of spain?",
			//"version":  1, this is automatically set by the DSL
		},
	})
	assert.NoError(t, err)

	//err = dsl.Update("question", map[string]interface{}{
	//	"sourceType": "brainly",
	//	"sourceId":   "1",
	//	"content":    "who was king of spain?",
	//	"version":    1, // this is required, and needs monotonically increasing version numbers
	//})
	//err = dsl.Archive("question", map[string]interface{}{
	//	"sourceType": "brainly",
	//	"sourceId":   "1",
	//	"content":    "who was king of spain?",
	//	//"version":  1, // version is not necessary
	//})
	//assert.NoError(t, err)

}
