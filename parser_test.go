// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsValidID(t *testing.T) {
	cases := []struct {
		id  string
		exp bool
	}{{
		id:  "a",
		exp: true,
	}, {
		id: "1",
	}}

	for _, c := range cases {
		got := isValidID(c.id)
		test.Assert(t, c.id, c.exp, got, true)
	}
}

func TestParseBlockAttribute(t *testing.T) {
	cases := []struct {
		in  string
		exp []string
	}{{
		in:  "",
		exp: nil,
	}, {
		in: "[]",
	}, {
		in: `[a]`,
		exp: []string{
			"a",
		},
	}, {
		in: `[a=2]`,
		exp: []string{
			"a=2",
		},
	}, {
		in: `[a=2,b="c, d",e,f=3]`,
		exp: []string{
			"a=2",
			`b=c, d`,
			"e",
			"f=3",
		},
	}, {
		in: `["A,B",w=_blank,role="a,b"]`,
		exp: []string{
			"A,B",
			"w=_blank",
			"role=a,b",
		},
	}}

	for _, c := range cases {
		got := parseBlockAttribute(c.in)
		test.Assert(t, "parseBlockAttribute", c.exp, got, true)
	}
}
