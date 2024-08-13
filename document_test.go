// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestOpen(t *testing.T) {
	var (
		doc  *Document
		fout *os.File
		err  error
	)

	doc, err = Open(`testdata/test.adoc`)
	if err != nil {
		t.Fatal(err)
	}

	// Since we cannot overwrite the asciidoctor output for
	// generator, we override ourself.
	doc.Attributes[DocAttrGenerator] = `Asciidoctor 2.0.18`

	fout, err = os.OpenFile(`testdata/test.got.html`, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = doc.ToHTML(fout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParse_document_title(t *testing.T) {
	type testCase struct {
		content   string
		expString string
		exp       DocumentTitle
	}

	var cases = []testCase{{
		content: `= Main: sub`,
		exp: DocumentTitle{
			Main: `Main`,
			Sub:  `sub`,
			sep:  defTitleSeparator,
		},
		expString: `Main: sub`,
	}, {
		// Without space after separator
		content: `= Main:sub`,
		exp: DocumentTitle{
			Main: `Main:sub`,
			sep:  defTitleSeparator,
		},
		expString: `Main:sub`,
	}, {
		// With multiple separator after separator
		content: `= a: b: c`,
		exp: DocumentTitle{
			Main: `a: b`,
			Sub:  `c`,
			sep:  defTitleSeparator,
		},
		expString: `a: b: c`,
	}, {
		// With custom separator.
		content: `
= Mainx sub
:title-separator: x`,
		exp: DocumentTitle{
			Main: `Main`,
			Sub:  `sub`,
			sep:  'x',
		},
		expString: `Mainx sub`,
	}}

	var (
		c   testCase
		got *Document
	)

	for _, c = range cases {
		got = Parse([]byte(c.content))
		test.Assert(t, `Main`, c.exp.Main, got.Title.Main)
		test.Assert(t, `Sub`, c.exp.Sub, got.Title.Sub)
		test.Assert(t, `sep`, c.exp.sep, got.Title.sep)
		test.Assert(t, `String`, c.expString, got.Title.String())
	}
}
