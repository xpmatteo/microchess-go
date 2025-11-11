
.PHONY: run test lint fmt tidy play-6502

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

GO6502_SOURCES := $(shell find go6502 -name "*.go" -not -name "*_test.go" -not -name "testrun.go")

go6502/go6502: $(GO6502_SOURCES)
	cd go6502 && go build -o go6502 main.go

go6502/microchess.bin: go6502/microchess.asm go6502/go6502
	cd go6502 && ./go6502 -a microchess.asm

play-6502: go6502/go6502 go6502/microchess.bin
	cd go6502 && ./go6502 -a microchess.asm && go run testrun.go iomem.go microchess.bin
