.SILENT:

debug:
	go run . run --debug

run:
	go run . run

test:
	go test ./...

build:
	go build . 

clean:
	rm hide

install:
	go install .

format:
	go fmt ./...

