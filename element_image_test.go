// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestElementParseBlockImage(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)
	tdata, err = test.LoadData(`testdata/element_image_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		caseName string
		input    []byte
		got      bytes.Buffer
	)
	for caseName, input = range tdata.Input {
		var doc = Parse(input)

		got.Reset()
		doc.ToHTMLEmbedded(&got)

		var exp = string(tdata.Output[caseName])
		test.Assert(t, caseName, exp, got.String())
	}
}
