// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseTableRows(t *testing.T) {
	cases := []struct {
		ncols   int
		lines   [][]byte
		expRow  tableRow
		expRest [][]byte
	}{{
		ncols: 0,
		lines: [][]byte{
			[]byte("| A"),
			[]byte("| B"),
		},
		expRow:  []string{"A", "B"},
		expRest: nil,
	}}

	for _, c := range cases {
		row, rest := parseTableRows(c.ncols, c.lines)
		test.Assert(t, "row", c.expRow, row, false)
		test.Assert(t, "rest", c.expRest, rest, false)
	}
}
