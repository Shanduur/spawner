.PHONY: run
run: build
	./build/spawner

.PHONY: build
build:
	go build -o ./build/spawner

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: tui
tui:
	go run examples/tui/main.go

.PHONY: iocp
iocp:
	go run examples/iocp/main.go
