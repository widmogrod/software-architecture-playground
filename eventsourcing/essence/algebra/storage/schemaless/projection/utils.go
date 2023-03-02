package schemaless

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func ConvertAs[A any](x schema.Schema) (A, error) {
	var a A
	var ret any
	var err error
	if any(a) == nil {
		ret, err = schema.ToGo(x)
	} else {
		ret, err = schema.ToGo(x, schema.WithExtraRules(schema.WhenPath(nil, schema.UseStruct(a))))
	}

	if err != nil {
		return a, err
	}

	result, ok := ret.(A)
	if !ok {
		return a, fmt.Errorf("cannot convert %T to %T", ret, a)
	}

	return result, nil
}

type ListAssert struct {
	t     *testing.T
	Items []Message
	Err   error
}

func (l *ListAssert) Returning(msg Message) {
	if l.Err != nil {
		panic(l.Err)
	}

	l.Items = append(l.Items, msg)
}

func (l *ListAssert) AssertLen(expected int) bool {
	return assert.Equal(l.t, expected, len(l.Items))
}

func (l *ListAssert) AssertAt(index int, expected Message) bool {
	return assert.Equal(l.t, expected, l.Items[index])
}

func (l *ListAssert) Contains(expected Message) bool {
	for _, item := range l.Items {
		if assert.Equal(l.t, expected, item) {
			return true
		}
	}

	l.t.Errorf("expected to find %v in result set but failed", expected)
	return false
}
