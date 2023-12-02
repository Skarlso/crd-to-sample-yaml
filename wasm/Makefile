build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build

run: build
	./wasm

static: build
	./wasm --static
