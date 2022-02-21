## SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all lint

all: lint test

lint:
	golangci-lint run ./...

test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

test-parser:
	asciidoctor --attribute stylesheet=empty.css testdata/test.adoc
	go test -v -run=Open .
