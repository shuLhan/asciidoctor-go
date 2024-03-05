// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseCellFormat(t *testing.T) {
	type testCase struct {
		exp *cellFormat
		raw string
	}

	var cases = []testCase{{
		raw: `3*`,
		exp: &cellFormat{
			ndupCol: 3,
		},
	}, {
		raw: `3+`,
		exp: &cellFormat{
			nspanCol: 3,
		},
	}, {
		raw: `.2+`,
		exp: &cellFormat{
			nspanRow: 2,
		},
	}, {
		raw: `2.3+`,
		exp: &cellFormat{
			nspanCol: 2,
			nspanRow: 3,
		},
	}, {
		raw: `^`,
		exp: &cellFormat{
			alignHor: colAlignMiddle,
		},
	}, {
		raw: `.<`,
		exp: &cellFormat{
			alignVer: colAlignTop,
		},
	}, {
		raw: `3+^.^`,
		exp: &cellFormat{
			nspanCol: 3,
			alignHor: colAlignMiddle,
			alignVer: colAlignMiddle,
		},
	}, {
		raw: `2*>m`,
		exp: &cellFormat{
			ndupCol:  2,
			alignHor: colAlignBottom,
			style:    colStyleMonospaced,
		},
	}, {
		raw: `.3+^.>s`,
		exp: &cellFormat{
			nspanRow: 3,
			alignHor: colAlignMiddle,
			alignVer: colAlignBottom,
			style:    colStyleStrong,
		},
	}, {
		raw: `.^l`,
		exp: &cellFormat{
			alignVer: colAlignMiddle,
			style:    colStyleLiteral,
		},
	}, {
		raw: `v`,
		exp: &cellFormat{
			style: colStyleVerse,
		},
	}}

	var (
		c   testCase
		got *cellFormat
	)

	for _, c = range cases {
		got = parseCellFormat(c.raw)
		test.Assert(t, c.raw, c.exp, got)
	}
}
