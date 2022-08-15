package di

import (
	"fmt"
	"testing"
)

type DIConfiger interface {
	ConfigGetString(string) string
	ConfigGetInt(string) int
	ConfigGetBool(string) bool
}

func LibBuilder(di DIConfiger) string {
	a := di.ConfigGetBool("enabled")
	b := di.ConfigGetString("db-conn")

	return fmt.Sprintf("lib-1(enabled=%t, db-conn=%s)", a, b)
}

func Lib2Builder(di DIConfiger) string {
	a := di.ConfigGetInt("workers")
	b := di.ConfigGetInt("timeout")

	return fmt.Sprintf("lib-2(workers=%d, timeout=%d)", a, b)
}

func Lib3Builder() string {
	return fmt.Sprintf("lib-3()")
}

type DICompontentBuilder interface {
	GetLib() string
	GetLib2() string
}

func ComponentBuilder(di DICompontentBuilder) string {
	a := di.GetLib()
	b := di.GetLib2()

	return fmt.Sprintf("component-1(lib-1=%s, lib-2=%s)", a, b)
}

type DICComponent2Builder interface {
	GetLib2() string
	GetLib3() string
}

func Component2Builder(di DICComponent2Builder) string {
	a := di.GetLib2()
	b := di.GetLib3()

	return fmt.Sprintf("component-2(lib-2=%s, lib-3=%s)", a, b)
}

type CIModuleBuilder interface {
	Component() string
	Component2() string
}

func ModuleBuilder(di CIModuleBuilder) string {
	a := di.Component()
	b := di.Component2()

	return fmt.Sprintf("module-1(%s, %s)", a, b)
}

type MyDI struct {
}

func (m *MyDI) ConfigGetString(s string) string {
	return "dummy"
}

func (m *MyDI) ConfigGetInt(s string) int {
	return 9999
}

func (m *MyDI) ConfigGetBool(s string) bool {
	return true
}

func (m *MyDI) GetLib() string {
	return LibBuilder(m)
}

func (m *MyDI) GetLib2() string {
	return Lib2Builder(m)
}

func (m *MyDI) GetLib3() string {
	return Lib3Builder()
}

func (m *MyDI) Component() string {
	return ComponentBuilder(m)
}

func (m *MyDI) Component2() string {
	return Component2Builder(m)
}

func (m *MyDI) GetModule() string {
	return ModuleBuilder(m)
}

func TestSimple(t *testing.T) {
	var ci MyDI

	module := ci.GetModule()
	t.Log(module)
}
