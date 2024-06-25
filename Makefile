.SILENT:

run:
	go run ./cmd/hide

test:
	go test ./...

build:
	go build ./cmd/hide

clean:
	rm hide

install:
	go install ./cmd/hide
