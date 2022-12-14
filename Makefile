## SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all lint test-parser serve-doc

all: test lint

lint:
	-golangci-lint run ./...
	-fieldalignment ./...

test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

test-parser:
	asciidoctor --attribute stylesheet=empty.css \
		--out-file=testdata/test.exp.html \
		testdata/test.adoc
	go test -v -run=Open .

serve-doc:
	ciigo serve _doc/
