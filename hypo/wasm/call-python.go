package main

import (
	"fmt"
	"github.com/wasmerio/wasmer-go/wasmer"
	"os"
)

func main() {
	// Let's declare the Wasm module.
	//
	// We are using the text representation of the module here.
	wasmBytes, err := os.ReadFile("./wapm_packages/rustpython/rustpython@0.1.3/target/wasm32-wasi/release/rustpython.wasm")
	if err != nil {
		panic(err)
	}

	// Create an Engine
	engine := wasmer.NewEngine()

	// Create a Store
	store := wasmer.NewStore(engine)

	fmt.Println("Compiling module...")
	module, err := wasmer.NewModule(store, wasmBytes)

	if err != nil {
		panic(fmt.Sprintln("Failed to compile module:", err))
	}

	wasiEnv, _ := wasmer.NewWasiStateBuilder("wasi-program").
		// Choose according to your actual situation
		Argument("python/main.py").
		// Environment("ABC", "DEF").
		MapDirectory("./", ".").
		Finalize()

	// Create an empty import object.
	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		panic(fmt.Sprintln("Failed to generate import object:", err))
	}

	importObject.Register("", map[string]wasmer.IntoExtern{
		"sumW": wasmer.NewFunction(
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

	start, err := instance.Exports.GetWasiStartFunction()
	if err != nil {
		panic(fmt.Sprintln("Failed to get start function:", err))
	}

	start()

	//sum, err := instance.Exports.GetRawFunction("sum")
	//
	//if err != nil {
	//	panic(fmt.Sprintln("Failed to retrieve the `sum` function:", err))
	//}
	//
	//fmt.Println("Calling `sum` function...")
	//// Let's call the `sum` exported function.
	//result, err := sum.Call(1, 2)
	//if err != nil {
	//	panic(fmt.Sprintln("Failed to call the `sum` function:", err))
	//}
	//
	//fmt.Println("Result of the `sum` function:", result)

	//instance.Exports.GetGlobal("memory").Set(wasmer.NewMemory(1, 1, false))

	//// Here we go.
	////
	//// The Wasm module exports a function called `sum`. Let's get
	//// it.
	//sum, err := instance.Exports.GetRawFunction("sum")
	//
	//if err != nil {
	//	panic(fmt.Sprintln("Failed to retrieve the `sum` function:", err))
	//}
	//
	//fmt.Println("Calling `sum` function...")
	//// Let's call the `sum` exported function.
	//result, err := sum.Call(1, 2)
	//
	//if err != nil {
	//	panic(fmt.Sprintln("Failed to call the `sum` function:", err))
	//}
	//
	//fmt.Println("Result of the `sum` function:", result)
	//
	//// That was fun. But what if we can get rid of the `Call` call? Well,
	//// that's possible with the native flavor. The function will seem like
	//// it's a standard Go function.
	//sumNative := sum.Native()
	//
	//fmt.Println("Calling `sum` function (natively)...")
	//// Let's call the `sum` exported function. The parameters are
	//// statically typed Rust values of type `i32` and `i32`. The
	//// result, in this case particular case, in a unit of type `i32`.
	//result, err = sumNative(3, 4)
	//
	//if err != nil {
	//	panic(fmt.Sprintln("Failed to call the `sum` function natively:", err))
	//}
	//
	//fmt.Println("Result of the `sum` function:", result)
}
