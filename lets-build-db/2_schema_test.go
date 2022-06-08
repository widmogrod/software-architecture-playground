package lets_build_db

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSchemaManagement(t *testing.T) {
	var schema = Schema{Fields: []Field{
		{Name: "id", Typ: TypeOf{Identifier: &Identifier{}}},
		{Name: "content", Typ: TypeOf{String: &String{}}},
		{Name: "created", Typ: TypeOf{DateTime: &DateTime{Format: "2006-01-02", DefaultNow: true}}},
	}}

	db := NewDB(schema)
	err, _ := db.InsertInto(map[string]interface{}{
		"content": "abcd",
		"created": time.Date(2022, 6, 5, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Equal(t, []KVSortedSet{
		{
			{"content:1", "abcd"},
			{"created:1", "2022-06-05"},
		},
	}, db.appendLog)
	err, _ = db.InsertInto(map[string]interface{}{
		"content": "abcd",
		"created": time.Date(2022, 6, 6, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Equal(t, []KVSortedSet{
		{
			{"content:1", "abcd"},
			{"created:1", "2022-06-05"},
		},
		{
			{"content:2", "abcd"},
			{"created:2", "2022-06-06"},
		},
	}, db.appendLog)

	result := db.Select(Predicate{Eq: &LeftOnly{
		Field: "content",
		Value: "abcd",
	}})
	assert.Equal(t, map[string]interface{}{
		"content": "abcd",
	}, result)
}
