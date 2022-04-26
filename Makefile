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
	go run tui/example/main.go
