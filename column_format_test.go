// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/math/big"
	"github.com/shuLhan/share/lib/test"
)

func TestParseColumnFormat(t *testing.T) {
	cases := []struct {
		s         string
		expNCols  int
		expFormat *columnFormat
	}{{
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

	for _, c := range cases {
		gotNCols, gotFormat := parseColumnFormat(c.s)
		test.Assert(t, c.s+" ncols", c.expNCols, gotNCols)
		test.Assert(t, c.s, c.expFormat, gotFormat)
	}
}
