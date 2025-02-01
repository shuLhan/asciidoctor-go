## SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all
all: test lint

.PHONY: lint
lint:
	go run ./internal/cmd/gocheck ./...
	go vet ./...

.PHONY: test
test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: test-parser
test-parser:
	asciidoctor --attribute stylesheet! \
		--out-file=testdata/test.exp.html \
		testdata/test.adoc
	go test -v -run=Open .

.PHONY: serve-doc
serve-doc:
	ciigo -address=127.0.0.1:31904 serve _doc/
