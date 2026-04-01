build:
	CGO_ENABLED=0 go build -o roundup ./cmd/roundup/

run: build
	./roundup

test:
	go test ./...

clean:
	rm -f roundup

.PHONY: build run test clean
