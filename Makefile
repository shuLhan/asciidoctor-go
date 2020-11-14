.PHONY: all lint

all: lint

lint:
	golangci-lint run ./...

test-parser:
	asciidoctor --attribute stylesheet=empty.css testdata/test.adoc
	go test -v -run=Open .
