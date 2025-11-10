
.PHONY: run test lint fmt tidy

run:
	go run ./cmd/microchess

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w pkg cmd acceptance

tidy:
	go mod tidy
