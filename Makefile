.PHONY: build test install clean

build:
	go build -o messages-relay ./cmd/messages-relay

test:
	go test ./...

install: build
	./scripts/install.sh

clean:
	rm -f messages-relay
