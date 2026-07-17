.PHONY: sync run build test clean

sync:
	CGO_ENABLED=0 go run ./cmd/bot sync

run:
	CGO_ENABLED=0 go run ./cmd/bot

build:
	CGO_ENABLED=0 go build -o telegram-career-bot ./cmd/bot

test:
	CGO_ENABLED=0 gotestsum --format testname ./...

clean:
	rm -rf data/ telegram-career-bot
