package main

import (
	"fmt"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func main() {
	// Let's declare the Wasm module.
	//
	// We are using the text representation of the module here.
	wasmBytes := []byte(`
		(module
          (import "env" "sum" (func $sum (param i32 i32) (result i32)))
		  (type $sum_t (func (param i32 i32) (result i32)))
		  (func $sum_f (type $sum_t) (param $x i32) (param $y i32) (result i32)
		    local.get $x
		    local.get $y
		    call $sum)
		  (export "sum" (func $sum_f)))
	`)

	// Create an Engine
	engine := wasmer.NewEngine()

	// Create a Store
	store := wasmer.NewStore(engine)

	fmt.Println("Compiling module...")
	module, err := wasmer.NewModule(store, wasmBytes)

	if err != nil {
		fmt.Println("Failed to compile module:", err)
	}

	// Create an empty import object.
	importObject := wasmer.NewImportObject()
	importObject.Register("env", map[string]wasmer.IntoExtern{
		"sum": wasmer.NewFunction(
			store,
			wasmer.NewFunctionType([]*wasmer.ValueType{
				wasmer.NewValueType(wasmer.I32), wasmer.NewValueType(wasmer.I32),
			}, []*wasmer.ValueType{
				wasmer.NewValueType(wasmer.I32),
			}),
			func(args []wasmer.Value) ([]wasmer.Value, error) {
				x := args[0].I32()
				y := args[1].I32()
				return []wasmer.Value{wasmer.NewI32(x + y)}, nil
			},
		),
	})

	fmt.Println("Instantiating module...")
	// Let's instantiate the Wasm module.
	instance, err := wasmer.NewInstance(module, importObject)

	if err != nil {
		panic(fmt.Sprintln("Failed to instantiate the module:", err))
	}

	// Here we go.
	//
	// The Wasm module exports a function called `sum`. Let's get
	// it.
	sum, err := instance.Exports.GetRawFunction("sum")

	if err != nil {
		panic(fmt.Sprintln("Failed to retrieve the `sum` function:", err))
	}

	fmt.Println("Calling `sum` function...")
	// Let's call the `sum` exported function.
	result, err := sum.Call(1, 2)

	if err != nil {
		panic(fmt.Sprintln("Failed to call the `sum` function:", err))
	}

	fmt.Println("Result of the `sum` function:", result)

	// That was fun. But what if we can get rid of the `Call` call? Well,
	// that's possible with the native flavor. The function will seem like
	// it's a standard Go function.
	sumNative := sum.Native()

	fmt.Println("Calling `sum` function (natively)...")
	// Let's call the `sum` exported function. The parameters are
	// statically typed Rust values of type `i32` and `i32`. The
	// result, in this case particular case, in a unit of type `i32`.
	result, err = sumNative(3, 4)

	if err != nil {
		panic(fmt.Sprintln("Failed to call the `sum` function natively:", err))
	}

	fmt.Println("Result of the `sum` function:", result)
}
