BINARY := yanzi-emitter

.PHONY: build run test unit

build:
	go build -o $(BINARY) ./cmd/yanzi-emitter

run: build
	./$(BINARY)

test:
	go test ./...

unit:
	go test ./...
