package lets_build_db

import (
	"encoding/json"
	"errors"
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
	return nil
}

func (s *DBValueSchema) Select(result Deserializable, p Predicate) (Identifiable, error) {
	res := Find(s.appendLog, func(kv KV) bool {
		if p.Eq != nil {
			key := kv[KEY]
			// Magic key ahead, consider doing it differently
			return p.Eq.Field == "@id" && key == p.Eq.Value
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
