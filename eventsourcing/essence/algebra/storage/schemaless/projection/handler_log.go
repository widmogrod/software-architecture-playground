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

func (l *LogHandler) Process(msg Message, returning func(Message)) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("%s: Combine(%s, %s) \n", l.prefix, x.Key, res)
			returning(msg)
			return nil
		},
		func(x *Retract) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("%s: Retract(%s, %s) \n", l.prefix, x.Key, res)
			returning(msg)
			return nil
		},
		func(x *Both) error {
			fmt.Printf("%s: Both(%s:\n", l.prefix, x.Key)
			fmt.Printf("\t")
			_ = l.Process(&x.Retract, returning)
			fmt.Printf("\t")
			_ = l.Process(&x.Combine, returning)
			fmt.Printf(")\n")
			returning(msg)
			return nil
		},
	)
}
