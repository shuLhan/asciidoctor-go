// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestOpen(t *testing.T) {
	var (
		doc *Document
		err error
	)
	doc, err = Open(`testdata/test.adoc`)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = doc.ToHTML(&buf)
	if err != nil {
		t.Fatal(err)
	}

	var redactLastUpdated = regexp.MustCompile(`Last updated (.*)`)
	var got = redactLastUpdated.ReplaceAll(buf.Bytes(),
		[]byte(`Last updated [REDACTED]`))

	err = os.WriteFile(`testdata/test.got.html`, got, 0600)
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

func TestDocumentSetAttribute(t *testing.T) {
	type testCase struct {
		desc     string
		key      string
		val      string
		expError string
		exp      DocumentAttribute
	}

	var doc = newDocument()
	clear(doc.Attributes.Entry)
	doc.Attributes.Entry[`key1`] = ``
	doc.Attributes.Entry[`key2`] = ``

	var listCase = []testCase{{
		key: `key3`,
		exp: doc.Attributes,
	}, {
		desc: `prefix negation`,
		key:  `!key1`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key2`: ``,
				`key3`: ``,
			},
		},
	}, {
		desc: `suffix negation`,
		key:  `key2!`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`: ``,
			},
		},
	}, {
		desc: `leveloffset +`,
		key:  docAttrLevelOffset,
		val:  `+2`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `+2`,
			},
			LevelOffset: 2,
		},
	}, {
		desc: `leveloffset -`,
		key:  docAttrLevelOffset,
		val:  `-2`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `-2`,
			},
			LevelOffset: 0,
		},
	}, {
		desc: `leveloffset`,
		key:  docAttrLevelOffset,
		val:  `1`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `1`,
			},
			LevelOffset: 1,
		},
	}, {
		desc:     `leveloffset: invalid`,
		key:      docAttrLevelOffset,
		val:      `*1`,
		expError: `Document: setAttribute: leveloffset invalid value "*1"`,
	}}

	var (
		tc  testCase
		err error
	)
	for _, tc = range listCase {
		err = doc.setAttribute(tc.key, tc.val)
		if err != nil {
			test.Assert(t, tc.desc+` error`, tc.expError, err.Error())
			continue
		}
		test.Assert(t, tc.desc, tc.exp, doc.Attributes)
	}
}
