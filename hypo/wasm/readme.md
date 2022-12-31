# wasm eperimetns

## setup wasm runtime
```
brew install rustup
rustup-init

brew install wasmtime
brew install wasmer
```


## compile golang to wasm
```bash
GOOS=js GOARCH=wasm go build -o main.wasm main.go
```

## run compiled code in wasm runtime
```
wasmtime main.wasm
Error: failed to run main module `main.wasm`

Caused by:
    0: failed to instantiate "main.wasm"
    1: unknown import: `go::debug` has not been defined
```

```
wasmer run main.wasm
error: failed to run `main.wasm`
╰─▶ 1: Error while importing "go"."debug": unknown import. Expected Function(FunctionType { params: [I32], results: [] })

```


-  problems is that golang runtime dont' support wasi (yet)! https://github.com/golang/go/issues/31105
-  to overcome it, we need to use tinygo https://tinygo.org/
- tinygo is a fork of golang, but with WASI support
- to compile tinygo, we need to install llvm and clang


## compile example using tinygo
Workds. But this has few constrains https://tinygo.org/docs/reference/lang-support/
```
brew tap tinygo-org/tools
brew install tinygo
```

```
tinygo build -wasm-abi=generic -target=wasi -o main-tiny.wasm  ./main.go
```

```
wasmtime main-tiny.wasm                                                 
Hello wasm
```

```
wasmer run main-tiny.wasm 
Hello wasm
```

## call wasm module from golang using wasmtime-go
and also call golang function from wasm!
```
go get -u github.com/bytecodealliance/wasmtime-go@v1.0.0

go run call-wasmtime.go
```

## call wasm module from golang using wasmer-go
and also call golang function from wasm!
```
go get -u github.com/wasmerio/wasmer-go@v1.0.4
go run call-wasmer.go
```

## call wasm modules from golang
### Install python interpreter in wasm
```
wapm install rustpython/rustpython
```

### run python in wasm
```
wasmer run rustpython/rustpython python/main.py --dir=.
```

### run python from golang as wasm module
```
go run call-python.go
```

