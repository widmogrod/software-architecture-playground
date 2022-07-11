package lets_build_db

import (
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNaiveDB(t *testing.T) {
	// In this db design schema management is not created

	//schema := [][3]string{
	//	{"question", "content", "string"},
	//	{"question", "creationDate", "string"},
	//}

	// - len to know when size is to big, and need to be split
	// - when there are split that will exist on disk,
	//   then it's important to know on which segment key exists
	appendLog := [][][2]string{
		{
			{"question:3:content", "content that I want to store 1"},
			{"question:3:creationDate", "2022-10-10"},
		},
	}
	assert.Len(t, appendLog, 1)

	appendLog = Set(appendLog, [][2]string{
		{"question:3:content", "content that I want to store 2"},
		{"question:3:author", "Gabriel"},
	})
	assert.Len(t, appendLog, 2)

	res := Get(appendLog, []string{"question:3:content", "question:3:creationDate"})
	assert.Equal(t, [][2]string{
		{"question:3:content", "content that I want to store 2"},
		{"question:3:creationDate", "2022-10-10"},
	}, res)

	compacted := Compact(Segment{}, appendLog)
	assert.Equal(t, []KVSortedSet{
		{
			{"question:3:content", "content that I want to store 2"},
			{"question:3:author", "Gabriel"},
			{"question:3:creationDate", "2022-10-10"},
		},
	}, compacted)

	res = Get(compacted, []string{"question:3:content", "question:3:creationDate"})
	assert.Equal(t, [][2]string{
		{"question:3:content", "content that I want to store 2"},
		{"question:3:creationDate", "2022-10-10"},
	}, res)

	res = Find(appendLog, func(kv KV) bool {
		return kv[VAL] == "Gabriel"
	}, 2)
	assert.Equal(t, [][2]string{
		{"question:3:author", "Gabriel"},
	}, res)
	//{`/question:\d+:content/`, "has", "store"},
	//{`/question:\d+:creationDate/`, "lt", "2023"},
	//}, 2) // Group operations would make sense

	t.Run("delete contract", func(t *testing.T) {
		appendLog = Delete(appendLog, []string{"question:3:content"})

		// deleted key MUST not be retrievable
		t.Run("deleted key MUST not be retrievable", func(t *testing.T) {
			res = Get(appendLog, []string{"question:3:content", "question:3:creationDate"})
			assert.Equal(t, [][2]string{
				{"question:3:creationDate", "2022-10-10"},
			}, res)
		})

		// deleted key MUST be comparable
		t.Run("deleted key MUST be comparable", func(t *testing.T) {
			compacted = Compact(Segment{}, appendLog)
			assert.Equal(t, []KVSortedSet{
				{
					{"question:3:author", "Gabriel"},
					{"question:3:creationDate", "2022-10-10"},
				},
			}, compacted)
		})

		// deleted key MUST not be findable
		t.Run("deleted key MUST not be findable", func(t *testing.T) {
			res = Find(appendLog, func(kv KV) bool {
				return kv[KEY] == "question:3:content"
			}, 2)
			assert.Equal(t, [][2]string(nil), res)
		})
	})
}

func TestRange(t *testing.T) {
	k := Pack("a", "b", "c")
	assert.Equal(t, []string{"a", "b", "c"}, k.Unpack())

	bt := btree.New(2)
	assert.False(t, bt.Has(k))

	i := bt.ReplaceOrInsert(k)
	assert.Nil(t, i)
	assert.True(t, bt.Has(k))

	j := bt.ReplaceOrInsert(k)
	assert.NotNil(t, j)
	assert.True(t, bt.Has(k))

	assert.Equal(t, bt.Get(k), j)

	bt.ReplaceOrInsert(Pack("b", "1"))
	bt.ReplaceOrInsert(Pack("b", "1", "x"))
	bt.ReplaceOrInsert(Pack("b", "1", "y"))
	bt.ReplaceOrInsert(Pack("b", "1", "z"))
	bt.ReplaceOrInsert(Pack("b", "10"))
	bt.ReplaceOrInsert(Pack("b", "2"))

	actual := []btree.Item{}
	expected := []btree.Item{
		Pack("b", "1", "x"),
		Pack("b", "1", "y"),
		Pack("b", "1", "z"),
	}
	bt.AscendRange(
		Pack("b", "1").Begin(),
		Pack("b", "1").End(),
		func(i btree.Item) bool {
			actual = append(actual, i)
			return true
		},
	)
	assert.Equal(t, expected, actual)
}
