package gm

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchemaGeneration(t *testing.T) {
	// new schema registry
	reg := NewSchemaRegistry()
	// register schema
	err := reg.Register("question", QuestionSchema())
	assert.NoError(t, err)

	obj := Question{}
	err = GenerateRandomData(reg, "question", &obj)
	assert.NoError(t, err)

	err = reg.Validate("question", obj)
	assert.NoError(t, err)

	// json serialize obj
	b, err := json.Marshal(obj)
	assert.NoError(t, err)
	fmt.Println(string(b))

	obj2 := map[string]interface{}{}
	err = GenerateRandomData(reg, "question", &obj2)
	assert.NoError(t, err)

	err = reg.Validate("question", obj2)
	assert.NoError(t, err)

	// json serialize obj
	b, err = json.Marshal(obj2)
	assert.NoError(t, err)
	fmt.Println(string(b))
}
