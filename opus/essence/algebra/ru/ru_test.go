package ru

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
)

func TestGenerateLamndaFromgRPC(t *testing.T) {
	assert.Equal(t, 1, 1)

	// new bytes buffer
	buf := bytes.NewBuffer([]byte(`
syntax = "proto3";
package question;

service QuestionService {
	rpc GetQuestion (GetQuestionRequest) returns (GetQuestionResponse) {}
}
`))

	dir := &DirectoryMock{
		MkdirFunc: func(name string, perm fs.FileMode) error {
			return nil
		},
		WriteFileFunc: func(name string, data []byte, perm fs.FileMode) error {
			return nil
		},
	}
	err := Suntetize(buf, dir)
	assert.NoError(t, err)

	// assert should call
	assert.Equal(t, "services/QuestionService/GetQuestion", dir.MkdirCalls()[0].Name)
	//assert.Equal(t, "services/QuestionService/GetQuestion/init.go", dir.WriteFileCalls()[0].Name)
}
