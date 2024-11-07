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

openapi:
	docker run -p 8081:8080 -v $(pwd):/tmp -e SWAGGER_FILE=/tmp/openapi.yaml swaggerapi/swagger-editor
