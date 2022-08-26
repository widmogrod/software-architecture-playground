package gm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPredicates(t *testing.T) {
	useCases := map[string]struct {
		predicate Predicate
		data      interface{}
		err       error
	}{
		"check if Typ accept correct type": {
			predicate: Predicate{Type: PtrType(TypeString)},
			data:      "asd",
		},
		"check if Typ reject incorrect type": {
			predicate: Predicate{Type: PtrType(TypeString)},
			data:      123123,
			err:       ErrWrongType,
		},
		"check if In predicate work as expected on correct data": {
			predicate: Predicate{
				In: []interface{}{
					"CreateRequest",
					"UpdateRequest",
					"ArchiveRequest",
				},
			},
			data: "CreateRequest",
		},
		"check if In predicate work as expected on incorrect data": {
			predicate: Predicate{
				In: []interface{}{
					"CreateRequest",
					"UpdateRequest",
					"ArchiveRequest",
				},
			},
			data: "DeleteRequest",
			err:  ErrValueNotContainedIn,
		},
		"check if Eq predicate work as expected on correct data": {
			predicate: Predicate{
				Eq: "CreateRequest",
			},
			data: "CreateRequest",
		},

		"check if Eq predicate work as expected on incorrect data": {
			predicate: Predicate{
				Eq: "CreateRequest",
			},
			data: "DeleteRequest",
			err:  ErrValueNotEqual,
		},
		"check if Fields predicate work as expected on correct data": {
			predicate: Predicate{
				Fields: map[string]Predicate{
					"action": {
						In: []interface{}{
							"CreateRequest",
							"UpdateRequest",
							"ArchiveRequest",
						},
					},
				},
			},
			data: map[string]interface{}{
				"action": "CreateRequest",
			},
		},
		"check if Fields predicate work as expected on incorrect data": {
			predicate: Predicate{
				Fields: map[string]Predicate{
					"action": {
						In: []interface{}{
							"CreateRequest",
							"UpdateRequest",
							"ArchiveRequest",
						},
					},
				},
			},
			data: map[string]interface{}{},
			err:  ErrFieldInMap,
		},
		"check if Fields predicate work as expected on incorrect data, string instead of map": {
			predicate: Predicate{
				Fields: map[string]Predicate{
					"action": {
						In: []interface{}{
							"CreateRequest",
							"UpdateRequest",
							"ArchiveRequest",
						},
					},
				},
			},
			data: "string instead of map",
			err:  ErrValueNotMap,
		},
		"check if And predicate work as expected on correct data": {
			predicate: Predicate{
				And: []Predicate{
					{Eq: "CreateRequest"},
					{In: []interface{}{"CreateRequest"}},
				},
			},
			data: "CreateRequest",
		},
		"check if And predicate work as expected on incorrect data": {
			predicate: Predicate{
				And: []Predicate{
					{Eq: "CreateRequest"},
					{In: []interface{}{"DeleteRequest"}},
				},
			},
			data: "CreateRequest",
			err:  ErrOneOfAndPredicatesFailed,
		},
		"check if Or predicate work as expected on correct data": {
			predicate: Predicate{
				Or: []Predicate{
					{Eq: "CreateRequest"},
					{Eq: "Some other request"},
					{In: []interface{}{"DeleteRequest"}},
				},
			},
			data: "DeleteRequest",
		},
		"check if Or predicate work as expected on incorrect data": {
			predicate: Predicate{
				Or: []Predicate{
					{Eq: "CreateRequest"},
					{Eq: "Some other request"},
					{In: []interface{}{"DeleteRequest"}},
				},
			},
			data: "this does not exist",
			err:  ErrAllOrPredicatesFailed,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			err := uc.predicate.Eval(uc.data)
			if err != nil {
				t.Log(err.Error())
				if uc.err != nil {
					assert.Contains(t, err.Error(), uc.err.Error())
				} else {
					assert.NoError(t, err, "error not expected but found")
				}
			} else {
				if uc.err != nil {
					assert.Error(t, uc.err, "error expected but not found")
				}
			}
		})
	}
}

func TestNewGuard(t *testing.T) {
	acl := NewGuard()
	err := acl.CreateRule("full-question-management", Predicate{Fields: map[string]Predicate{
		"action": {In: []interface{}{
			"CreateRequest",
			"UpdateRequest",
			"ArchiveRequest",
		}},
	}})
	assert.NoError(t, err)

	err = acl.CreateRule("full-question-management", Predicate{Eq: "asd"})
	if assert.Error(t, err) {
		assert.ErrorContains(t, err, ErrRuleAlreadyIdExists.Error())
	}

	err = acl.CreteRuleBaseOf("limited-question-management", "full-question-management", Predicate{Fields: map[string]Predicate{
		"payload": {Fields: map[string]Predicate{
			"data": {Fields: map[string]Predicate{
				"sourceType": {Eq: "quora.com"},
			}},
		}},
	}})
	assert.NoError(t, err)

	useCases := map[string]struct {
		rule RuleID
		data interface{}
		err  error
	}{
		"non existing rule should return error": {
			rule: "my-rule-id",
			err:  ErrRuleNotFound,
		},
		"check if action is allowed": {
			rule: "limited-question-management",
			data: map[string]interface{}{
				"action": "CreateRequest",
				"payload": map[string]interface{}{
					"data": map[string]interface{}{
						"sourceType": "quora.com",
					},
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			err := acl.EvalRule(uc.rule, uc.data)
			if err != nil {
				t.Log(err.Error())
				if uc.err != nil {
					assert.Contains(t, err.Error(), uc.err.Error())
				} else {
					assert.NoError(t, err, "error not expected but found")
				}
			} else {
				if uc.err != nil {
					assert.Error(t, uc.err, "error expected but not found")
				}
			}
		})
	}
}
