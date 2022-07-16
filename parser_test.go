// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestGenerateID(t *testing.T) {
	type testCase struct {
		inputExp map[string]string
		doc      *Document
		desc     string
	}

	var cases = []testCase{{
		desc: `Without idprefix and idseparator`,
		doc: &Document{
			Attributes: AttributeEntry{},
		},
		inputExp: map[string]string{
			`a b c`:  `a_b_c`,
			` a b c`: `_a_b_c`,
		},
	}, {
		desc: `With idprefix`,
		doc: &Document{
			Attributes: AttributeEntry{
				metaNameIDPrefix: `123`,
			},
		},
		inputExp: map[string]string{
			`a b c`:  `_123a_b_c`,
			` a b c`: `_123_a_b_c`,
		},
	}, {
		desc: `With empty idseparator`,
		doc: &Document{
			Attributes: AttributeEntry{
				metaNameIDSeparator: ``,
			},
		},
		inputExp: map[string]string{
			`a b c`:  `abc`,
			` a b c`: `abc`,
		},
	}, {
		desc: `With idseparator`,
		doc: &Document{
			Attributes: AttributeEntry{
				metaNameIDSeparator: `-`,
			},
		},
		inputExp: map[string]string{
			`a b c`:  `a-b-c`,
			` a b c`: `_-a-b-c`,
		},
	}, {
		desc: `With idprefix and idseparator`,
		doc: &Document{
			Attributes: AttributeEntry{
				metaNameIDPrefix:    `id_`,
				metaNameIDSeparator: `-`,
			},
		},
		inputExp: map[string]string{
			`a b c`:  `id_a-b-c`,
			` a b c`: `id_-a-b-c`,
		},
	}}

	var (
		c   testCase
		inp string
		exp string
		got string
	)

	for _, c = range cases {
		for inp, exp = range c.inputExp {
			got = generateID(c.doc, inp)
			test.Assert(t, c.desc, exp, got)
		}
	}
}

func TestIsValidID(t *testing.T) {
	type testCase struct {
		id  string
		exp bool
	}

	var cases = []testCase{{
		id:  "a",
		exp: true,
	}, {
		id: "1",
	}}

	var (
		c   testCase
		got bool
	)

	for _, c = range cases {
		got = isValidID([]byte(c.id))
		test.Assert(t, c.id, c.exp, got)
	}
}
