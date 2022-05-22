package lets_build_db

import (
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
			{"question:3:content", "content that I want to store"},
			{"question:3:creationDate", "2022-10-10"},
		},
	}

	Set(appendLog, [][2]string{
		{"question:3:content", "content that I want to store"},
		{"question:3:author", "Gabriel"},
	})

	res := Get(appendLog, []string{"question:3:content", "question:3:creationDate"})
	assert.Equal(t, [][2]string{
		{"question:3:content", "content that I want to store"},
		{"question:3:creationDate", "2022-10-10"},
	}, res)

	//Find(appendLog, [][3]string{
	//	{`/question:\d+:content/`, "has", "store"},
	//	{`/question:\d+:creationDate/`, "lt", "2023"},
	//}, 2) // Group operations would make sense

	//result := Find(schema, appendLog, "select * from schema where asd_asd_asd_ADs = '123'")
}
