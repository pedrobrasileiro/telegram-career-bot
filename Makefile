.PHONY: sync run build clean

sync:
	CGO_ENABLED=0 go run . sync

run:
	CGO_ENABLED=0 go run .

build:
	CGO_ENABLED=0 go build -o telegram-career-bot .

clean:
	rm -rf data/ telegram-career-bot
