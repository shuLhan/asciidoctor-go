// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParserInline_do(t *testing.T) {
	doc := &Document{}
	tmpl, err := doc.createHTMLTemplate()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		content string
		exp     string
	}{{
		content: "__A__B",
		exp:     "<em>A</em>B",
	}, {
		content: "__A *B*__",
		exp:     "<em>A <strong>B</strong></em>",
	}, {
		content: "__A _B_ C__",
		exp:     "<em>A <em>B</em> C</em>",
	}, {
		content: "__A B_ C__",
		exp:     "<em>A B_ C</em>",
	}, {
		content: "__A *B*_",
		exp:     "<em>_A <strong>B</strong></em>",
	}, {
		content: "_A *B*__",
		exp:     "<em>A <strong>B</strong>_</em>",
	}, {
		content: "_A_B",
		exp:     "_A_B",
	}, {
		content: "_A_ B",
		exp:     "<em>A</em> B",
	}, {
		content: "_A _B",
		exp:     "_A _B",
	}, {
		content: "*A*B",
		exp:     "*A*B",
	}, {
		content: "*A* B",
		exp:     "<strong>A</strong> B",
	}, {
		content: "*A *B",
		exp:     "*A *B",
	}, {
		content: "`A`B",
		exp:     "`A`B",
	}, {
		content: "`A` B",
		exp:     "<code>A</code> B",
	}, {
		content: "`A `B",
		exp:     "`A `B",
	}, {
		content: "`+__A *B*__+`",
		exp:     "<code>__A *B*__</code>",
	}, {
		content: "`++__A *B*__++`",
		exp:     "<code>__A *B*__</code>",
	}, {
		content: "`++__A *B*__+`",
		exp:     "<code>+__A *B*__</code>",
	}, {
		content: "*A _B `C_ D` E*",
		exp:     "<strong>A <em>B <code>C</code></em><code> D</code> E</strong>",
	}, {
		content: "A bold with * space *, with single non alnum *=*.",
		exp:     "A bold with * space <strong>, with single non alnum *=</strong>.",
	}, {
		content: "*bold _italic `mono end-bold* end-italic_ end-mono.",
		exp:     "<strong>bold <em>italic `mono end-bold</em></strong><em> end-italic</em> end-mono.",
	}, {
		content: "\"`A double quote without end.",
		exp:     "\"`A double quote without end.",
	}, {
		content: "\"` A double quote around space `\"",
		exp:     "\"` A double quote around space `\"",
	}, {
		content: "\"`A double quote`\"",
		exp:     "&#8220;A double quote&#8221;",
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err = container.toHTML(doc, tmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}
