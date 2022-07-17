// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDocumentParser_parseHeader(t *testing.T) {
	type testCase struct {
		expDoc      func() *Document
		desc        string
		content     string
		expPreamble string
	}

	var cases = []testCase{{
		desc:    "With empty line contains white spaces",
		content: "//// block\ncomment.\n////\n\t  \n= Title\n",
		expDoc: func() (doc *Document) {
			doc = newDocument()
			return doc
		},
		expPreamble: "= Title",
	}}

	var (
		c      testCase
		expDoc *Document
		gotDoc *Document
	)

	for _, c = range cases {
		t.Log(c.desc)

		expDoc = c.expDoc()
		gotDoc = Parse([]byte(c.content))

		test.Assert(t, "Title", expDoc.Title.raw, gotDoc.Title.raw)
		test.Assert(t, "rawAuthors", expDoc.rawAuthors, gotDoc.rawAuthors)
		test.Assert(t, "rawRevision", expDoc.rawRevision, gotDoc.rawRevision)
		test.Assert(t, "Attributes", expDoc.Attributes, gotDoc.Attributes)
		test.Assert(t, "Preamble text", c.expPreamble, gotDoc.preamble.toText())
	}
}

func TestDocumentParser_parseListDescription_withOpenBlock(t *testing.T) {
	var content = []byte(`
Description:: Description body with open block.
+
--
Paragraph A.

* List item 1
* List item 2
--

Paragraph C.
`)

	var exp string = `
<div class="dlist">
<dl>
<dt class="hdlist1">Description</dt>
<dd>
<p>Description body with open block.</p>
<div class="openblock">
<div class="content">
<div class="paragraph">
<p>Paragraph A.</p>
</div>
<div class="ulist">
<ul>
<li>
<p>List item 1</p>
</li>
<li>
<p>List item 2</p>
</li>
</ul>
</div>
</div>
</div>
</dd>
</dl>
</div>
<div class="paragraph">
<p>Paragraph C.</p>
</div>`

	var (
		doc *Document = Parse(content)

		got bytes.Buffer
		err error
	)

	err = doc.ToHTMLEmbedded(&got)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "parseListDescription with open block", exp, got.String())
}
