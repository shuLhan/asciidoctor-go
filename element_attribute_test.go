// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func Test_parseElementAttribute(t *testing.T) {
	type testCase struct {
		raw string
		exp elementAttribute
	}

	var cases = []testCase{{
		raw: ``,
	}, {
		raw: `[]`,
	}, {
		raw: `[STYLE]`,
		exp: elementAttribute{
			rawStyle: `STYLE`,
		},
	}, {
		raw: `[style#id]`,
		exp: elementAttribute{
			ID:       `id`,
			rawStyle: `style`,
		},
	}, {
		raw: `[#id.role1.role2,options="opt1,opt2"]`,
		exp: elementAttribute{
			ID:      `id`,
			roles:   []string{`role1`, `role2`},
			options: []string{`opt1`, `opt2`},
			pos:     1,
		},
	}, {
		raw: `[cols="3*,^"]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameCols: `3*,^`,
			},
		},
	}, {
		raw: `[quote, attribution]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameAttribution: `attribution`,
			},
			rawStyle: `quote`,
			style:    styleQuote,
			pos:      1,
		},
	}, {
		raw: `[quote, attribution, citation]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameAttribution: `attribution`,
				attrNameCitation:    `citation`,
			},
			rawStyle: `quote`,
			style:    styleQuote,
			pos:      2,
		},
	}}

	var (
		c testCase
	)

	for _, c = range cases {
		var got = elementAttribute{}
		got.parseElementAttribute([]byte(c.raw))
		test.Assert(t, c.raw, c.exp, got)
	}
}
