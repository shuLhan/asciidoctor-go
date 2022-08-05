// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/test"
)

func TestInlineParser_do(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: "*A _B `C_ D` E*",
		exp:     "<strong>A <em>B <code>C</code></em><code> D</code> E</strong>",
	}, {
		content: "A * B *, C *=*.",
		exp:     "A * B <strong>, C *=</strong>.",
	}, {
		content: "*A _B `C D* E_ F.",
		exp:     "<strong>A <em>B `C D</em></strong><em> E</em> F.",
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)
	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseAttrRef(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		Attributes: map[string]string{
			`x`: `https://kilabit.info`,
		},
	}

	var cases = []testCase{{
		content: `A {x}[*B*] C`,
		exp:     `A <a href="https://kilabit.info"><strong>B</strong></a> C`,
	}, {
		content: `A {x }[*B*] C`,
		exp:     `A <a href="https://kilabit.info"><strong>B</strong></a> C`,
	}, {
		content: `A {x }*B* C`,
		exp:     `A <a href="https://kilabit.info*B*" class="bare">https://kilabit.info*B*</a> C`,
	}, {
		content: `A {y }*B* C`,
		exp:     `A {y }<strong>B</strong> C`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseCrossReference(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: map[string]*anchor{
			`x`: &anchor{
				label: `X y`,
			},
		},
		titleID: map[string]string{
			`X y`: `x`,
		},
	}

	var cases = []testCase{{
		content: `A <<x>>`,
		exp:     `A <a href="#x">X y</a>`,
	}, {
		content: `A <<x, Label>>`,
		exp:     `A <a href="#x">Label</a>`,
	}, {
		content: `A <<X y>>`,
		exp:     `A <a href="#x">X y</a>`,
	}, {
		content: `A <<X y,Label>>`,
		exp:     `A <a href="#x">Label</a>`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)
	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseFormat(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `_A_B`,
		exp:     `_A_B`,
	}, {
		content: `_A_ B`,
		exp:     `<em>A</em> B`,
	}, {
		content: `_A _B`,
		exp:     `_A _B`,
	}, {
		content: `*A*B`,
		exp:     `*A*B`,
	}, {
		content: `*A* B`,
		exp:     `<strong>A</strong> B`,
	}, {
		content: `*A *B`,
		exp:     `*A *B`,
	}, {
		content: "`A`B",
		exp:     "`A`B",
	}, {
		content: "`A` B",
		exp:     `<code>A</code> B`,
	}, {
		content: "`A `B",
		exp:     "`A `B",
	}, {
		content: "A `/**/` *B*",
		exp:     "A <code>/<strong></strong>/</code> <strong>B</strong>",
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseFormatUnconstrained(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `__A__B`,
		exp:     `<em>A</em>B`,
	}, {
		content: `__A *B*__`,
		exp:     `<em>A <strong>B</strong></em>`,
	}, {
		content: `__A _B_ C__`,
		exp:     `<em>A <em>B</em> C</em>`,
	}, {
		content: `__A B_ C__`,
		exp:     `<em>A B_ C</em>`,
	}, {
		content: `__A *B*_`,
		exp:     `<em>_A <strong>B</strong></em>`,
	}, {
		content: `_A *B*__`,
		exp:     `<em>A <strong>B</strong>_</em>`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseInlineID(t *testing.T) {
	type testCase struct {
		content  string
		exp      string
		isForToC bool
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `[[A]] B`,
		exp:     `<a id="A"></a> B`,
	}, {
		content:  `[[A]] B`,
		exp:      ` B`,
		isForToC: true,
	}, {
		content: `[[A] B`,
		exp:     `[[A] B`,
	}, {
		content: `[A]] B`,
		exp:     `[A]] B`,
	}, {
		content: `[[A ]] B`,
		exp:     `[[A ]] B`,
	}, {
		content: `[[ A]] B`,
		exp:     `[[ A]] B`,
	}, {
		content: `[[A B]] C`,
		exp:     `[[A B]] C`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		_testDoc.isForToC = c.isForToC
		container.toHTML(_testDoc, &buf)
		_testDoc.isForToC = false

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseInlineIDShort(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `[#A]#B#`,
		exp:     `<span id="A">B</span>`,
	}, {
		content: `[#A]#B`,
		exp:     `[#A]#B`,
	}, {
		content: `[#A]B#`,
		exp:     `[#A]B#`,
	}, {
		content: `[#A ]#B#`,
		exp:     `[#A ]#B#`,
	}, {
		content: `[# A]# B#`,
		exp:     `[# A]# B#`,
	}, {
		content: `[#A B]# C#`,
		exp:     `[#A B]# C#`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseInlineImage(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `image:https://upload.wikimedia.org/wikipedia/commons/3/35/Tux.svg[Linux,25,35]`,
		exp:     `<span class="image"><img src="https://upload.wikimedia.org/wikipedia/commons/3/35/Tux.svg" alt="Linux" width="25" height="35"></span>`,
	}, {
		content: `image:linux.png[Linux,150,150,float="right"]
You can find Linux everywhere these days!`,
		exp: `<span class="image right"><img src="linux.png" alt="Linux" width="150" height="150"></span>
You can find Linux everywhere these days!`,
	}, {
		content: `image:sunset.jpg[Sunset,150,150,role="right"] What a beautiful sunset!`,
		exp:     `<span class="image right"><img src="sunset.jpg" alt="Sunset" width="150" height="150"></span> What a beautiful sunset!`,
	}, {
		content: `image:sunset.jpg[Sunset]
image:linux.png[2]`,
		exp: `<span class="image"><img src="sunset.jpg" alt="Sunset"></span>
<span class="image"><img src="linux.png" alt="2"></span>`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parsePassthrough(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
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

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parsePassthroughDouble(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
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

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parsePassthroughTriple(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
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

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseQuote(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
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

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)
	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseSubscsript(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
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

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseSuperscript(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `A^B^C`,
		exp:     `A<sup>B</sup>C`,
	}, {
		content: `A^B ^C`,
		exp:     `A^B ^C`,
	}, {
		content: `A^ B^C`,
		exp:     `A^ B^C`,
	}, {
		content: `A\^B^C`,
		exp:     `A^B^C`,
	}, {
		content: `A^B\^C`,
		exp:     `A^B^C`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}

func TestInlineParser_parseURL(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors: make(map[string]*anchor),
		titleID: make(map[string]string),
	}

	var cases = []testCase{{
		content: `https://asciidoctor.org/abc`,
		exp:     `<a href="https://asciidoctor.org/abc" class="bare">https://asciidoctor.org/abc</a>`,
	}, {
		content: `https://asciidoctor.org.`,
		exp:     `<a href="https://asciidoctor.org" class="bare">https://asciidoctor.org</a>.`,
	}, {
		content: `https://asciidoctor.org[Asciidoctor^,role="a,b"].`,
		exp:     `<a href="https://asciidoctor.org" class="a b" target="_blank" rel="noopener">Asciidoctor</a>.`,
	}, {
		content: `\https://example.org.`,
		exp:     `https://example.org.`,
	}, {
		content: `irc://irc.freenode.org/#fedora[Fedora IRC channel].`,
		exp:     `<a href="irc://irc.freenode.org/#fedora">Fedora IRC channel</a>.`,
	}, {
		content: `mailto:ms@kilabit.info.`,
		exp:     `<a href="mailto:ms@kilabit.info">mailto:ms@kilabit.info</a>.`,
	}, {
		content: `mailto:ms@kilabit.info[Mail to me].`,
		exp:     `<a href="mailto:ms@kilabit.info">Mail to me</a>.`,
	}, {
		content: `Relative file link:test.html[test.html].`,
		exp:     `Relative file <a href="test.html">test.html</a>.`,
	}, {
		content: `link:https://kilabit.info[Kilabit^].`,
		exp:     `<a href="https://kilabit.info" target="_blank" rel="noopener">Kilabit</a>.`,
	}, {
		content: `http: this is not link`,
		exp:     `http: this is not link`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		if debug.Value > 0 {
			container.debug(0)
		}

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}
