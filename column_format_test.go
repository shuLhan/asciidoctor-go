// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/math/big"
	"github.com/shuLhan/share/lib/test"
)

func TestParseColumnFormat(t *testing.T) {
	type testCase struct {
		expFormat *columnFormat
		s         string
		expNCols  int
	}

	var cases = []testCase{{
		s:        "3*",
		expNCols: 3,
		expFormat: &columnFormat{
			isDefault: true,
			width:     big.NewRat(1),
		},
	}, {
		s:        "3*^",
		expNCols: 3,
		expFormat: &columnFormat{
			alignHor:  colAlignMiddle,
			isDefault: true,
			width:     big.NewRat(1),
		},
	}, {
		s:        "3*.^",
		expNCols: 3,
		expFormat: &columnFormat{
			alignVer:  colAlignMiddle,
			isDefault: true,
			width:     big.NewRat(1),
		},
	}}

	var (
		c         testCase
		gotNCols  int
		gotFormat *columnFormat
	)

	for _, c = range cases {
		gotNCols, gotFormat = parseColumnFormat(c.s)
		test.Assert(t, c.s+" ncols", c.expNCols, gotNCols)
		test.Assert(t, c.s, c.expFormat, gotFormat)
	}
}
