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
<div id="header">
<h1>Abc</h1>
<div class="details">
</div>
</div>
<div id="content">
<div id="preamble">
<div class="sectionbody">
<div class="paragraph">
<p>Abc begins on a bleary Monday morning.</p>
</div>
</div>
</div>
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
		err := doc.ToHTMLBody(&buf)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, "", expHTML, buf.String(), true)
	}
}
