package deserializer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

func NewDeSerializer() *Deserializer {
	return &Deserializer{
		store: sync.Map{},
	}
}

type Deserializer struct {
	store sync.Map
}

func (d *Deserializer) Register(typ interface{}) {
	d.RegisterName(d.name(typ), typ)
}

func (d *Deserializer) RegisterName(name string, typ interface{}) {
	rt := reflect.TypeOf(typ)
	if rt.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("RegisterName. Registred type must not be a pointer, but given %T", typ))
	}

	actual, loaded := d.store.LoadOrStore(name, rt)
	if loaded {
		panic(fmt.Sprintf("RegisterName. Type %T was already registered under name %s, cannot reguster twice.", actual, name))
	}
}

func (d *Deserializer) Name(typ interface{}) (string, error) {
	name := d.name(typ)
	if _, found := d.store.Load(name); found {
		return name, nil
	}

	return "", fmt.Errorf("type %T is not in register", typ)
}

func (d *Deserializer) name(typ interface{}) string {
	rt := reflect.TypeOf(typ)

	// Remove information about pointer
	name := strings.TrimLeft(rt.String(), "*")
	if name == "" {
		panic(fmt.Sprintf("Accept only named types, but given %T ", typ))
	}

	return name
}

func (d *Deserializer) Serialise(typ interface{}) ([]byte, error) {
	name := d.name(typ)
	if _, found := d.store.Load(name); found {
		return json.Marshal(typ)
	}

	return nil, fmt.Errorf("Serialise. Try to serialise type %T that is not registered.", typ)
}

func (d *Deserializer) DeSerialiseName(name string, data []byte) (interface{}, error) {
	if val, found := d.store.Load(name); found {
		rt := val.(reflect.Type)
		value := reflect.New(rt)
		res := value.Interface()
		return res, json.Unmarshal(data, res)
	}

	return nil, fmt.Errorf("DeSerialiseName. Try to deserialise type with name %s, but is not registered", name)
}
