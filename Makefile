GOROOT := $(shell go env GOROOT)
WASM_EXEC_JS := $(GOROOT)/lib/wasm/wasm_exec.js

.PHONY: all wasm server clean

## all: build WASM module then the server binary (default target)
all: wasm server

## wasm: compile the Go WASM module and copy wasm_exec.js into static/
wasm:
	GOARCH=wasm GOOS=js go build -o static/main.wasm ./wasm/
	cp $(WASM_EXEC_JS) static/wasm_exec.js

## server: build the Go HTTP server binary
server:
	go build -o gistio .

## run: build everything and start the server
run: all
	./gistio

## clean: remove build artifacts
clean:
	rm -f static/main.wasm static/wasm_exec.js gistio
