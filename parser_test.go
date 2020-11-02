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

func TestParseAttributeElement(t *testing.T) {
	cases := []struct {
		in       string
		expKey   string
		expValue string
		expOpts  []string
	}{{
		in: "",
	}, {
		in: "[]",
	}, {
		in:     `[a]`,
		expKey: "a",
		expOpts: []string{
			"a",
		},
	}, {
		in:       `[a=2]`,
		expKey:   "a",
		expValue: "2",
		expOpts: []string{
			"a=2",
		},
	}, {
		in:       `[a=2,b="c, d",e,f=3]`,
		expKey:   "a",
		expValue: "2",
		expOpts: []string{
			"a=2",
			`b=c, d`,
			"e",
			"f=3",
		},
	}, {
		in:     `["A,B",w=_blank,role="a,b"]`,
		expKey: "A,B",
		expOpts: []string{
			"A,B",
			"w=_blank",
			"role=a,b",
		},
	}}

	for _, c := range cases {
		key, val, opts := parseAttributeElement(c.in)
		test.Assert(t, "attribute name", c.expKey, key, true)
		test.Assert(t, "attribute value", c.expValue, val, true)
		test.Assert(t, "opts", c.expOpts, opts, true)
	}
}
