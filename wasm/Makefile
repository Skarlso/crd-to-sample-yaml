build:
	GOARCH=wasm GOOS=js go build -ldflags "-X main.baseURL=http://localhost:8000" -o web/app.wasm
	go build -ldflags "-X main.baseURL=http://localhost:8000"

build-static:
	GOARCH=wasm GOOS=js go build -ldflags "-X main.baseURL=https://crdtoyaml.com" -o web/app.wasm
	go build -ldflags "-X main.baseURL=https://crdtoyaml.com"

run: build
	./wasm

static: build-static
	./wasm --static
