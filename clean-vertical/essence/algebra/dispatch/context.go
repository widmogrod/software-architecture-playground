package dispatch

import (
	"context"
	"github.com/segmentio/ksuid"
	"sync"
)

type Context interface {
	ActivityID() string
	Ctx() context.Context
	Set(key, value interface{})
	Get(key interface{}) (interface{}, bool)
}

func Background() *activity {
	return WithContext(context.Background())
}

func FromActivityID(id string) *activity {
	return &activity{
		params:     &sync.Map{},
		activityID: id,
		ctx:        context.Background(),
	}
}

func WithContext(ctx context.Context) *activity {
	uuid, err := ksuid.NewRandom()
	if err != nil {
		panic("dispatch: cannot generate activity ID. reason: " + err.Error())
	}

	return &activity{
		params:     &sync.Map{},
		activityID: uuid.String(),
		ctx:        ctx,
	}
}

type activity struct {
	params     *sync.Map
	activityID string
	ctx        context.Context
}

func (ctx *activity) ActivityID() string {
	return ctx.activityID
}

func (ctx *activity) Ctx() context.Context {
	return ctx.ctx
}

func (ctx *activity) Set(key, value interface{}) {
	ctx.params.Store(key, value)
}

func (ctx *activity) Get(key interface{}) (interface{}, bool) {
	return ctx.params.Load(key)
}
