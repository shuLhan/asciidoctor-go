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
	cases := []struct {
		content string
		exp     string
	}{{
		content: "*A _B `C_ D` E*",
		exp:     "<strong>A <em>B <code>C</code></em><code> D</code> E</strong>",
	}, {
		content: "A * B *, C *=*.",
		exp:     "A * B <strong>, C *=</strong>.",
	}, {
		content: "*A _B `C D* E_ F.",
		exp:     "<strong>A <em>B `C D</em></strong><em> E</em> F.",
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parseFormat(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
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
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}

}

func TestParserInline_parseFormatUnconstrained(t *testing.T) {
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
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parsePassthrough(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: "`+__A *B*__+`",
		exp:     "<code>__A *B*__</code>",
	}, {
		content: `\+__A *B*__+`,
		exp:     `+<em>A <strong>B</strong></em>+`,
	}, {
		content: `+__A *B*__\+`,
		exp:     `+<em>A <strong>B</strong></em>+`,
	}, {
		content: `X+__A *B*__+`,
		exp:     `X+<em>A <strong>B</strong></em>+`,
	}, {
		content: `+__A *B*__+X`,
		exp:     `+<em>A <strong>B</strong></em>+X`,
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parsePassthroughDouble(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: "`++__A *B*__++`",
		exp:     "<code>__A *B*__</code>",
	}, {
		content: "`++__A *B*__+`",
		exp:     "<code><em>A <strong>B</strong></em>+</code>",
	}, {
		content: `\++__A *B*__++`,
		exp:     `+__A *B*__+`,
	}, {
		content: `+\+__A *B*__++`,
		exp:     `+__A *B*__+`,
	}, {
		content: `++__A *B*__\++`,
		exp:     `<em>A <strong>B</strong></em>++`,
	}, {
		content: `++__A *B*__+\+`,
		exp:     `<em>A <strong>B</strong></em>++`,
	}, {
		content: `++ <u>A</u> ++`,
		exp:     ` <u>A</u> `,
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parsePassthroughTriple(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: `+++__A *B*__+++`,
		exp:     `__A *B*__`,
	}, {
		content: `+++__A *B*__++`,
		exp:     `+__A *B*__`,
	}, {
		content: `\+++__A *B*__+++`,
		exp:     `+__A *B*__+`,
	}, {
		content: `+\++__A *B*__+++`,
		exp:     `+<em>A <strong>B</strong></em>+`,
	}, {
		content: `++\+__A *B*__+++`,
		exp:     `+__A *B*__+`,
	}, {
		content: `+++__A *B*__\+++`,
		exp:     `+__A *B*__+`,
	}, {
		content: `+++__A *B*__+\++`,
		exp:     `__A *B*__++`,
	}, {
		content: `+++__A *B*__++\+`,
		exp:     `+__A *B*__+`,
	}, {
		content: `+++ <u>A</u> +++`,
		exp:     ` <u>A</u> `,
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parseQuote(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: "\"`A double quote without end.",
		exp:     "\"`A double quote without end.",
	}, {
		content: "\"` A double quote around space `\"",
		exp:     "\"` A double quote around space `\"",
	}, {
		content: "\"`A double quote`\"",
		exp:     "&#8220;A double quote&#8221;",
	}, {
		content: "\"`Escaped double quote\\`\"",
		exp:     "\"`Escaped double quote`\"",
	}, {
		content: "'`A single quote without end.",
		exp:     "'`A single quote without end.",
	}, {
		content: "'` A single quote around space `'",
		exp:     "'` A single quote around space &#8217;",
	}, {
		content: "\"`A single quote`\"",
		exp:     "&#8220;A single quote&#8221;",
	}, {
		content: "\"`Escaped single quote\\`\"",
		exp:     "\"`Escaped single quote`\"",
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parseSubscsript(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: "A~B~C",
		exp:     "A<sub>B</sub>C",
	}, {
		content: "A~B ~C",
		exp:     "A~B ~C",
	}, {
		content: "A~ B~C",
		exp:     "A~ B~C",
	}, {
		content: `A\~B~C`,
		exp:     "A~B~C",
	}, {
		content: `A~B\~C`,
		exp:     "A~B~C",
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}

func TestParserInline_parseSuperscript(t *testing.T) {
	cases := []struct {
		content string
		exp     string
	}{{
		content: "A^B^C",
		exp:     "A<sup>B</sup>C",
	}, {
		content: "A^B ^C",
		exp:     "A^B ^C",
	}, {
		content: "A^ B^C",
		exp:     "A^ B^C",
	}, {
		content: `A\^B^C`,
		exp:     "A^B^C",
	}, {
		content: `A^B\^C`,
		exp:     "A^B^C",
	}}

	var buf bytes.Buffer
	for _, c := range cases {
		buf.Reset()

		container := parseInlineMarkup([]byte(c.content))
		err := container.toHTML(_testDoc, _testTmpl, &buf)
		if err != nil {
			t.Fatal(err)
		}

		// container.debug(0)

		got := buf.String()
		test.Assert(t, c.content, c.exp, got, true)
	}
}
