package gm

import (
	"errors"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
)

type Config struct {
	tenantId string
}

type DSL struct {
	reg   *SchemaRegistry
	conf  *Config
	acl   *Guard
	store *kv.Store
}

type CreateQuestionRequest struct {
	Id      string `json:"sourceId"`
	Content string `json:"content"`
}

type Attributes = map[string]interface{}

var ErrAccessDenied = errors.New("access denied")

type InvokeRequest struct {
	Action  string      `name:"action"`
	Payload interface{} `name:"payload"`
}

func (d *DSL) Invoke(in *CreateQuestionRequest) error {
	request := InvokeRequest{
		Action:  "CreateQuestionRequest",
		Payload: in,
	}

	err := d.acl.EvalRule(d.conf.tenantId, request)
	if err != nil {
		// wrap error with details
		return fmt.Errorf("invoke1: %s: %w", err, ErrAccessDenied)
	}

	data := Question{
		Content:    in.Content,
		SourceId:   in.Id,
		SourceType: d.conf.tenantId,
		Version:    1,
	}

	err = d.reg.Validate("question", data)
	if err != nil {
		return fmt.Errorf("invoke2: %s: %w", err, ErrAccessDenied)
	}

	err = d.store.SetAttributes(data.ToKey(), data.ToAttr())
	if err != nil {
		return fmt.Errorf("invoke3: %s: %w", err, ErrAccessDenied)
	}

	return nil
}

func (d *DSL) Update(entity string, data Attributes) error {
	return errors.New("not implemented")
}

func (d *DSL) Archive(entity string, data Attributes) error {
	return errors.New("not implemented")
}

func NewStorageDSL(reg *SchemaRegistry, acl *Guard, store *kv.Store, c *Config) *DSL {
	return &DSL{
		reg:   reg,
		conf:  c,
		acl:   acl,
		store: store,
	}
}
