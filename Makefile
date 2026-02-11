BINARY := yanzi-emitter

.PHONY: build run

build:
	go build -o $(BINARY) ./cmd/yanzi-emitter

run: build
	./$(BINARY)
