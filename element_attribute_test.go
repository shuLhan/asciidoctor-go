// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func Test_parseElementAttribute(t *testing.T) {
	cases := []struct {
		raw string
		exp elementAttribute
	}{{
		raw: "",
	}, {
		raw: "[]",
	}, {
		raw: "[STYLE]",
		exp: elementAttribute{
			rawStyle: "STYLE",
		},
	}, {
		raw: "[style#id]",
		exp: elementAttribute{
			ID:       "id",
			rawStyle: "style",
		},
	}, {
		raw: `[#id.role1.role2,options="opt1,opt2"]`,
		exp: elementAttribute{
			ID:      "id",
			roles:   []string{"role1", "role2"},
			options: []string{"opt1", "opt2"},
			pos:     1,
		},
	}, {
		raw: `[cols="3*,^"]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameCols: "3*,^",
			},
		},
	}, {
		raw: `[quote, attribution]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameAttribution: "attribution",
			},
			rawStyle: "quote",
			style:    styleQuote,
			pos:      1,
		},
	}, {
		raw: `[quote, attribution, citation]`,
		exp: elementAttribute{
			Attrs: map[string]string{
				attrNameAttribution: "attribution",
				attrNameCitation:    "citation",
			},
			rawStyle: "quote",
			style:    styleQuote,
			pos:      2,
		},
	}}

	for _, c := range cases {
		got := elementAttribute{}
		got.parseElementAttribute([]byte(c.raw))
		test.Assert(t, c.raw, c.exp, got)
	}
}
