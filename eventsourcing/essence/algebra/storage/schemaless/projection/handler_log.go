package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

type LogHandler struct{}

func Log() Handler {
	return &LogHandler{}
}

func (l *LogHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("Log: Combine(%s, %s) \n", x.Key, res)
			return returning(msg)
		},
		func(x *Retract) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("Log: Retract(%s, %s) \n", x.Key, res)
			return returning(msg)

		},
		func(x *Both) error {
			fmt.Printf("Log: Both(%s:\n", x.Key)
			fmt.Printf("\t")
			_ = l.Process(&x.Retract, returning)
			fmt.Printf("\t")
			_ = l.Process(&x.Combine, returning)
			fmt.Printf(") Both end\n")
			return returning(msg)
		},
	)
}
