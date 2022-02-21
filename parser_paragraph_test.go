// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParser_parseParagraph(t *testing.T) {
	cases := []struct {
		desc    string
		content []byte
		exp     string
	}{{
		desc: "with lead style",
		content: []byte(`[.lead]
This is the ultimate paragraph.`),
		exp: `
<div class="paragraph lead">
<p>This is the ultimate paragraph.</p>
</div>`,
	}}

	parentDoc := newDocument()
	out := bytes.Buffer{}

	for _, c := range cases {
		subdoc := parseSub(parentDoc, c.content)
		out.Reset()
		err := subdoc.ToHTMLEmbedded(&out)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.desc, c.exp, out.String())
	}
}
