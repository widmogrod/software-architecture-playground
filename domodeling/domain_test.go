package domodeling

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Create struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestParseDomain(t *testing.T) {
	body := []byte(`{"name":"Sam", "age": 100}`)
	shape := Create{}
	_ = json.Unmarshal(body, &shape)

	dom := ParseDomain(shape)
	int := InterpreterInMem()
	result := ExecuteDomain(dom, int)

	response, _ := json.Marshal(result)
	assert.JSONEq(t, `{"create-result":{"success":"dom:user:666"}}`, string(response))
}

func TestParseDomain2(t *testing.T) {
	body := []byte(`{"create":{"name":"Sam", "age": 100}}`)
	dom := DomAST{}
	_ = json.Unmarshal(body, &dom)

	int := InterpreterInMem()
	result := ExecuteDomain(dom, int)

	response, _ := json.Marshal(result)
	assert.JSONEq(t, `{"create-result":{"success":"dom:user:666"}}`, string(response))
}

type (
	ruid = string
	name = string
	age  = int

	DomAST struct {
		Create *CreateAST `json:"create"`
	}

	CreateAST struct {
		Name name
		Age  age
	}
)

type CreateRes struct {
	Success ruid `json:"success"`
}

type Interpret interface {
	Create(ast CreateAST) CreateRes
}

type Result struct {
	CreateResult *CreateRes `json:"create-result,omitempty"`
}

func ParseDomain(shape Create) DomAST {
	return DomAST{
		Create: &CreateAST{
			Name: shape.Name,
			Age:  MkAge(shape.Age),
		},
	}
}

func MkAge(a int) age {
	if a < 10 || a > 130 {
		panic(fmt.Sprintf("age: out of bounds %d < 10 and %d > 130", a, a))
	}

	return a
}

func InterpreterInMem() Interpret {
	return &InMemInterpreter{}
}

var _ Interpret = &InMemInterpreter{}

type InMemInterpreter struct {
}

func (i *InMemInterpreter) Create(ast CreateAST) CreateRes {
	fmt.Printf("call Create(%v)\n", ast)

	return CreateRes{
		Success: "dom:user:666",
	}
}

func ExecuteDomain(dom DomAST, int Interpret) Result {
	if dom.Create != nil {
		res := int.Create(*dom.Create)
		return Result{
			CreateResult: &res,
		}
	}

	panic("don't know")
}
