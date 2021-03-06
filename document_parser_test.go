// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParse_metaDocTitle(t *testing.T) {
	expHTML := `
<div class="paragraph">
<p>Abc begins on a bleary Monday morning.</p>
</div>`

	cases := []struct {
		content string
	}{{
		content: `= Abc

{doctitle} begins on a bleary Monday morning.`,
	}, {
		content: `:doctitle: Abc

{doctitle} begins on a bleary Monday morning.`,
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		doc := Parse([]byte(c.content))
		buf.Reset()
		err := doc.ToHTMLEmbedded(&buf)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, "", expHTML, buf.String())
	}
}

func TestParse_metaShowTitle(t *testing.T) {
	cases := []struct {
		desc    string
		content string
		expHTML string
	}{{
		desc:    "default",
		content: `= Abc`,
		expHTML: `
<div id="header">
<h1>Abc</h1>
<div class="details">
</div>
</div>
<div id="content">
<div id="preamble">
<div class="sectionbody">
</div>
</div>
</div>
<div id="footer">
<div id="footer-text">
</div>
</div>`,
	}, {
		desc: "with showtitle!",
		content: `= Abc
:showtitle!:`,
		expHTML: `
<div id="header">
<div class="details">
</div>
</div>
<div id="content">
<div id="preamble">
<div class="sectionbody">
</div>
</div>
</div>
<div id="footer">
<div id="footer-text">
</div>
</div>`,
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		doc := Parse([]byte(c.content))
		buf.Reset()
		err := doc.ToHTMLBody(&buf)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.desc, c.expHTML, buf.String())
	}
}

func TestParse_document_title(t *testing.T) {
	cases := []struct {
		content   string
		exp       DocumentTitle
		expString string
	}{{
		content: `= Main: sub`,
		exp: DocumentTitle{
			Main: "Main",
			Sub:  "sub",
			sep:  defTitleSeparator,
		},
		expString: "Main: sub",
	}, {
		// Without space after separator
		content: `= Main:sub`,
		exp: DocumentTitle{
			Main: "Main:sub",
			sep:  defTitleSeparator,
		},
		expString: "Main:sub",
	}, {
		// With multiple separator after separator
		content: `= a: b: c`,
		exp: DocumentTitle{
			Main: "a: b",
			Sub:  "c",
			sep:  defTitleSeparator,
		},
		expString: "a: b: c",
	}, {
		// With custom separator.
		content: `:title-separator: x
= Mainx sub`,
		exp: DocumentTitle{
			Main: "Main",
			Sub:  "sub",
			sep:  'x',
		},
		expString: "Mainx sub",
	}}

	for _, c := range cases {
		got := Parse([]byte(c.content))
		test.Assert(t, "Main", c.exp.Main, got.Title.Main)
		test.Assert(t, "Sub", c.exp.Sub, got.Title.Sub)
		test.Assert(t, "sep", c.exp.sep, got.Title.sep)
		test.Assert(t, "String", c.expString, got.Title.String())
	}
}

func TestParse_author(t *testing.T) {
	cases := []struct {
		desc    string
		content string
		exp     []*Author
	}{{
		desc: "single author",
		content: `= T
A B`,
		exp: []*Author{{
			FirstName: "A",
			LastName:  "B",
			Initials:  "AB",
		}},
	}, {
		desc: "single author with email",
		content: `= T
A B <a@b>`,
		exp: []*Author{{
			FirstName: "A",
			LastName:  "B",
			Initials:  "AB",
			Email:     "a@b",
		}},
	}, {
		desc: "multiple authors",
		content: `= T
A B <a@b>; C <c@c>; D e_f G <>;`,
		exp: []*Author{{
			FirstName: "A",
			LastName:  "B",
			Initials:  "AB",
			Email:     "a@b",
		}, {
			FirstName: "C",
			Initials:  "C",
			Email:     "c@c",
		}, {
			FirstName:  "D",
			MiddleName: "e f",
			LastName:   "G",
			Initials:   "DeG",
		}},
	}, {
		desc: "meta author",
		content: `= T
:author: A B
:email: a@b`,
		exp: []*Author{{
			FirstName: "A",
			LastName:  "B",
			Initials:  "AB",
			Email:     "a@b",
		}},
	}}

	for _, c := range cases {
		got := Parse([]byte(c.content))
		test.Assert(t, c.desc, c.exp, got.Authors)
	}
}

func TestDocumentParser_parseListDescription_withOpenBlock(t *testing.T) {
	content := []byte(`
Description:: Description body with open block.
+
--
Paragraph A.

* List item 1
* List item 2
--

Paragraph C.
`)

	exp := `
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

	doc := Parse(content)
	got := bytes.Buffer{}

	err := doc.ToHTMLEmbedded(&got)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "parseListDescription with open block", exp, got.String())
}
