package lets_build_db

import (
	"encoding/json"
	"errors"
	"fmt"
)

func NewDbValueSchema() *DBValueSchema {
	return &DBValueSchema{}
}

type (
	DBValueSchema struct {
		appendLog AppendLog
	}
	Identifiable interface {
		Identifier() string
	}
	Serializable interface {
		Identifier() string
		Serialize() (string, error)
	}
	Deserializable interface {
		Identifier() string
		Deserialize([]byte) (Identifiable, error)
	}

	Predicate2 struct {
		Path  []string
		Value Value
	}
	Value struct {
		Eq interface{}
		//And *PredicateBool
	}
	//PredicateBool struct {
	//	Left  Predicate2
	//	Right Predicate2
	//}
)

func (s *DBValueSchema) Insert(record Serializable) error {
	key := record.Identifier()
	val, err := record.Serialize()
	if err != nil {
		return err
	}
	s.appendLog = Set(s.appendLog, KVSortedSet{
		{key, val},
	})

	// TODO index creation

	return nil
}

// Rename Select to lookup
// Create DSL that search only on indexed fields
func (s *DBValueSchema) Select(result Deserializable, p Predicate2) (Identifiable, error) {
	res := Find(s.appendLog, func(kv KV) bool {
		if p.Path[0] == "@id" {
			key := kv[KEY]
			// Magic key ahead, consider doing it differently
			return key == p.Value.Eq
		}
		return false
	}, 1)

	if len(res) == 0 {
		return nil, errors.New("no results")
	}

	// TODO streaming of the result
	return result.Deserialize([]byte(res[0][VAL]))
}

var _ Identifiable = &Json{}
var _ Serializable = &Json{}
var _ Deserializable = &Json{}

func FromJson(i Identifiable) *Json {
	return &Json{i: i}
}

type Json struct {
	i Identifiable
}

func (j *Json) Identifier() string {
	return j.i.Identifier()
}

func (j *Json) Serialize() (string, error) {
	res, err := json.Marshal(j.i)
	return string(res), err
}

func (j *Json) Deserialize(data []byte) (Identifiable, error) {
	r := j.i
	return r, json.Unmarshal(data, r)
}

type Person struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (p *Person) Identifier() string {
	return p.Id
}

// function that adds two numbers
func Add(a, b int) int {
	return a + b
}

// test Add function from above
func TestAdd() {
	fmt.Println(Add(1, 2))
}

// function that read buffer and maps values by key
func Map(buf []byte, fn func(string, string) string) string {
	var m map[string]string
	json.Unmarshal(buf, &m)
	for k, v := range m {
		m[k] = fn(k, v)
	}
	res, _ := json.Marshal(m)
	return string(res)
}
