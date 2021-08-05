package stream

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestChannelStream(t *testing.T) {
	s := NewChannelStream()
	go s.Work()

	size := 10
	for range rand.Perm(size) {
		s.Push(*GenerateMessage())
	}

	Spec(t, s, &specC{
		batchSize: size,
	})
}

func TestChannelStream_SelectOnce(t *testing.T) {
	messages := []Message{
		{
			Kind: "test",
		},
		{
			Kind: "test2",
			Data: []byte(`{"a":123}`),
		},
		{
			Kind: "test2",
		},
		{
			Kind: "test",
			Data: []byte(`{"b":{"a":321}}`),
		},
	}

	useCases := map[string]struct {
		given        []Message
		selectOnce   SelectOnceCMD
		expectResult []*Message
	}{
		"should return match when Kind is specified and match": {
			given: messages,
			selectOnce: SelectOnceCMD{
				Kind:         "test",
				MaxFetchSize: 1,
			},
			expectResult: []*Message{
				{
					Kind: "test",
				},
			},
		},
		"should match one element when MaxFetchSize is 0?": {
			given: messages,
			selectOnce: SelectOnceCMD{
				Kind:         "test",
				MaxFetchSize: 0,
			},
			expectResult: []*Message{
				{
					Kind: "test",
				},
			},
		},
		"should return matching result with KeyExists": {
			given: messages,
			selectOnce: SelectOnceCMD{
				Kind: "test2",
				Selector: &SelectConditions{
					KeyExists: "a",
				},
			},
			expectResult: []*Message{
				{
					Kind: "test2",
					Data: []byte(`{"a":123}`),
				},
			},
		},
		"should return result when Key `a` has value eq to `123`": {
			given: messages,
			selectOnce: SelectOnceCMD{
				Kind: "test2",
				Selector: &SelectConditions{
					KeyValue: map[string]SelectConditions{
						"a": {Eq: float64(123)},
					},
				},
			},
			expectResult: []*Message{
				{
					Kind: "test2",
					Data: []byte(`{"a":123}`),
				},
			},
		},
		"select should work for nested results": {
			given: messages,
			selectOnce: SelectOnceCMD{
				Kind: "test",
				Selector: &SelectConditions{
					KeyValue: map[string]SelectConditions{
						"b": {
							KeyValue: map[string]SelectConditions{
								"a": {Eq: float64(321)},
							},
						},
					},
				},
				MaxFetchSize: 1,
			},
			expectResult: []*Message{
				{
					Kind: "test",
					Data: []byte(`{"b":{"a":321}}`),
				},
			},
		},
	}

	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			s := NewChannelStream()
			for _, m := range uc.given {
				s.Push(m)
			}

			go func() {
				time.Sleep(time.Millisecond * 500)
				s.Work()
			}()

			result := s.SelectOnce(uc.selectOnce)
			assert.ElementsMatch(t, uc.expectResult, result)
		})
	}
}

type specC struct {
	batchSize int
}

func Spec(t *testing.T, s Streamer, c *specC) {
	t.Run("Push", func(t *testing.T) {
		for range rand.Perm(c.batchSize) {
			s.Push(*GenerateMessage())
		}
	})
	t.Run("Fetch cannot return more than batch size", func(t *testing.T) {
		result := s.Fetch(c.batchSize)
		assert.Truef(t, c.batchSize >= len(result), "c.batchSize = %d && len(result) = %d", c.batchSize, len(result))
		assert.True(t, c.batchSize >= 1)
	})
}
