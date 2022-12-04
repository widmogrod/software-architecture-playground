package ru

import (
	"bytes"
	"github.com/emicklei/proto"
	"os"
	"path"
)

//go:generate moq -out ru_moq_test.go . Directory

type Directory interface {
	Mkdir(name string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
}

type Conf struct {
	Directory string
}

func (c *Conf) LambdaDir(service, method string) string {
	return path.Join(c.Directory, service, method)
}

type Visitor struct {
	proto.NoopVisitor
	conf        *Conf
	serviceName string
	dir         Directory
}

func (v *Visitor) VisitService(service *proto.Service) {
	v.serviceName = service.Name
	for _, e := range service.Elements {
		e.Accept(v)
	}
}

func (v *Visitor) VisitRPC(rpc *proto.RPC) {
	err := v.dir.Mkdir(v.conf.LambdaDir(v.serviceName, rpc.Name), 0755)
	if err != nil {
		panic(err)
	}
}

func Suntetize(buf *bytes.Buffer, dir Directory) error {
	conf := &Conf{
		Directory: "services",
	}

	parser := proto.NewParser(buf)
	definition, err := parser.Parse()
	if err != nil {
		return err
	}

	visitor := &Visitor{
		conf: conf,
		dir:  dir,
	}
	definition.Accept(visitor)

	//proto.Walk(proto.WithService(func(service *proto.Service) {
	//			service.Accept(proto.WithRPC(func(rpc *proto.RPC) {
	//
	//			}))
	//			for _, method := range service.Methods {
	//				dir.Mkdir(conf.LambdaDir(service, method), 0755)
	//				dir.WriteFile(method.Name+"/init.go", []byte(method.Name), 0644)
	//			}
	//		}),
	//	definition,
	//	proto.WithRPC(func(rpc *proto.RPC) {
	//		rpc.Parent.Accept()
	//	}))

	return nil
}
