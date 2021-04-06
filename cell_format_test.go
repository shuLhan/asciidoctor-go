// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseCellFormat(t *testing.T) {
	cases := []struct {
		raw string
		exp *cellFormat
	}{{
		raw: "3*",
		exp: &cellFormat{
			ndupCol: 3,
		},
	}, {
		raw: "3+",
		exp: &cellFormat{
			nspanCol: 3,
		},
	}, {
		raw: ".2+",
		exp: &cellFormat{
			nspanRow: 2,
		},
	}, {
		raw: "2.3+",
		exp: &cellFormat{
			nspanCol: 2,
			nspanRow: 3,
		},
	}, {
		raw: "^",
		exp: &cellFormat{
			alignHor: colAlignMiddle,
		},
	}, {
		raw: ".<",
		exp: &cellFormat{
			alignVer: colAlignTop,
		},
	}, {
		raw: "3+^.^",
		exp: &cellFormat{
			nspanCol: 3,
			alignHor: colAlignMiddle,
			alignVer: colAlignMiddle,
		},
	}, {
		raw: "2*>m",
		exp: &cellFormat{
			ndupCol:  2,
			alignHor: colAlignBottom,
			style:    colStyleMonospaced,
		},
	}, {
		raw: ".3+^.>s",
		exp: &cellFormat{
			nspanRow: 3,
			alignHor: colAlignMiddle,
			alignVer: colAlignBottom,
			style:    colStyleStrong,
		},
	}, {
		raw: ".^l",
		exp: &cellFormat{
			alignVer: colAlignMiddle,
			style:    colStyleLiteral,
		},
	}, {
		raw: "v",
		exp: &cellFormat{
			style: colStyleVerse,
		},
	}}

	for _, c := range cases {
		got := parseCellFormat(c.raw)
		test.Assert(t, c.raw, c.exp, got)
	}
}
