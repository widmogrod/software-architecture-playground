package lets_build_db

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDocDB(t *testing.T) {
	doc := MapAny{
		"$docId":    "id123",
		"to-Leave?": true,
		"some":      9.2,
		"other": ListAny{
			"a", "b", "c",
			ListAny{"x", "y", "z"},
			MapAny{
				"nested": "123",
				"list":   ListAny{"1x", "1y", "1z"},
			},
		},
	}

	json1, err := json.Marshal(doc)
	assert.NoError(t, err)
	fmt.Println(string(json1))

	json2, err := json.Marshal(Unflatten(Flatten(doc, Pack())))
	assert.NoError(t, err)
	fmt.Println(string(json2))

	assert.JSONEq(t, string(json1), string(json2))

	db := NewDocDb()
	result, err := db.Save(doc)
	assert.NoError(t, err)
	json3, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.JSONEq(t, string(json1), string(json3))

	result, err = db.Get(result["$docId"].(string))
	assert.NoError(t, err)
	json4, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.JSONEq(t, string(json1), string(json4))

	for i := 0; i < 100; i++ {
		delete(doc, "$docId")
		result1, err := db.Save(doc)
		assert.NoError(t, err)
		result2, err := db.Get(result1["$docId"].(string))
		assert.NoError(t, err)

		json1, err := json.Marshal(result1)
		assert.NoError(t, err)
		json2, err := json.Marshal(result2)
		assert.NoError(t, err)
		assert.JSONEq(t, string(json1), string(json2))
	}
}
