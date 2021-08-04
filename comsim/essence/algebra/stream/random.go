package stream

import (
	"github.com/segmentio/ksuid"
)

var _ Streamer = &RandomStream{}

func NewRandomStream() *RandomStream {
	return &RandomStream{}
}

type RandomStream struct {
}

func (p *RandomStream) Push(message Message) {
	// do nothing
}

func (p *RandomStream) Fetch(size int) []*Message {
	result := make([]*Message, 0, size)
	for i := 0; i < size; i++ {
		result = append(result, GenerateMessage())
	}
	return result
}

func GenerateMessage() *Message {
	return &Message{
		Data: ksuid.New().Bytes(),
	}
}
