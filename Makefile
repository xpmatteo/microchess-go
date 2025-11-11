
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

play-6502:
	
	cd go6502 && ./go6502 -a microchess.asm && go run testrun.go iomem.go microchess.bin
