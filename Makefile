## SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all lint test-parser test-doctype-man serve-doc

all: test lint

lint:
	-fieldalignment ./...
	-shadow ./...
	-golangci-lint run \
		--presets bugs,metalinter,performance,unused \
		./...

test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

test-parser:
	asciidoctor --attribute stylesheet! \
		--out-file=testdata/test.exp.html \
		testdata/test.adoc
	go test -v -run=Open .

test-doctype-man:
	asciidoctor \
		--backend=manpage \
		--out-file=testdata/man_backend/test.exp.7 \
		testdata/man_backend/test.adoc

serve-doc:
	ciigo serve _doc/
