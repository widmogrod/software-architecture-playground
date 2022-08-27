package gm

import (
	"errors"
	"fmt"
)

type Config struct {
	tenantId string
}

type DSL struct {
	reg  *SchemaRegistry
	conf *Config
	acl  *Guard
}

type CreateRequest struct {
	Entity string
	Data   map[string]interface{}
}

type Attributes = map[string]interface{}

var ErrAccessDenied = errors.New("access denied")

type InvokeRequest struct {
	Action  string      `name:"action"`
	Payload interface{} `name:"payload"`
}

func (d *DSL) Invoke(in *CreateRequest) error {
	request := InvokeRequest{
		Action:  "CreateRequest",
		Payload: in,
	}

	err := d.acl.EvalRule(d.conf.tenantId, request)
	if err != nil {
		// wrap error with details
		return fmt.Errorf("invoke: %s: %w", err, ErrAccessDenied)
	}

	return errors.New("not implemented")
}

func (d *DSL) Update(entity string, data Attributes) error {
	return errors.New("not implemented")
}

func (d *DSL) Archive(entity string, data Attributes) error {
	return errors.New("not implemented")
}

func NewStorageDSL(reg *SchemaRegistry, acl *Guard, c *Config) *DSL {
	return &DSL{
		reg:  reg,
		conf: c,
		acl:  acl,
	}
}
