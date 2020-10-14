.PHONY: all lint install serve build build-release

all: install

lint:
	golangci-lint run ./...

test-parser:
	asciidoctor --attribute stylesheet=empty.css testdata/test.adoc
	go test -v -run=Open .
