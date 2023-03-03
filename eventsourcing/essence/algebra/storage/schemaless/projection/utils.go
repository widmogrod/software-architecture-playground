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

func Each(x schema.Schema, f func(value schema.Schema)) {
	_ = schema.MustMatchSchema(
		x,
		func(x *schema.None) any {
			return nil
		},
		func(x *schema.Bool) any {
			f(x)
			return nil
		},
		func(x *schema.Number) any {
			f(x)
			return nil
		},
		func(x *schema.String) any {
			f(x)
			return nil
		},
		func(x *schema.Binary) any {
			f(x)
			return nil
		},
		func(x *schema.List) any {
			for _, v := range x.Items {
				f(v)
			}
			return nil
		},
		func(x *schema.Map) any {
			f(x)
			return nil
		},
	)
}

type ListAssert struct {
	t     *testing.T
	Items []Item
	Err   error
}

func (l *ListAssert) Returning(msg Item) {
	if l.Err != nil {
		panic(l.Err)
	}

	l.Items = append(l.Items, msg)
}

func (l *ListAssert) AssertLen(expected int) bool {
	return assert.Equal(l.t, expected, len(l.Items))
}

func (l *ListAssert) AssertAt(index int, expected Item) bool {
	return assert.Equal(l.t, expected, l.Items[index])
}

func (l *ListAssert) Contains(expected Item) bool {
	for _, item := range l.Items {
		if assert.Equal(l.t, expected, item) {
			return true
		}
	}

	l.t.Errorf("expected to find %v in result set but failed", expected)
	return false
}
