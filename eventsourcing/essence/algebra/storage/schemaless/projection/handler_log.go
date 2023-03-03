package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

type LogHandler struct {
	prefix string
}

func Log(prefix string) Handler {
	return &LogHandler{
		prefix: prefix,
	}
}

func (l *LogHandler) Process(x Item, returning func(Item)) error {
	res, err := schema.ToJSON(x.Data)
	if err != nil {
		return err
	}
	fmt.Printf("%s: Item(%s, %s) \n", l.prefix, x.Key, res)
	returning(x)
	return nil
}

func (l *LogHandler) Retract(x Item, returning func(Item)) error {
	//TODO implement me
	panic("implement me")
}
