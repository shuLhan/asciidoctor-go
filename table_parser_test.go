// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestTableParser_new(t *testing.T) {
	cases := []struct {
		desc    string
		content string
		exp     []*tableCell
	}{{
		desc:    "empty content",
		content: ``,
		exp:     nil,
	}, {
		desc:    "first cell without |",
		content: "A1|B1",
		exp: []*tableCell{{
			content: []byte("A1"),
		}, {
			content: []byte("B1"),
		}},
	}, {
		desc:    "first cell without |",
		content: "A1\nb|B1",
		exp: []*tableCell{{
			content: []byte("A1\nb"),
		}, {
			content: []byte("B1"),
		}},
	}, {
		desc:    "single row",
		content: `|A1|B1`,
		exp: []*tableCell{{
			content: []byte("A1"),
		}, {
			content: []byte("B1"),
		}},
	}, {
		desc:    "two rows, empty header",
		content: "\n|A1",
		exp: []*tableCell{nil, {
			content: []byte("A1"),
		}},
	}, {
		desc:    "three rows, empty header",
		content: "\n|A1 |\n\nb\n\n|A2",
		exp: []*tableCell{nil, {
			content: []byte("A1 "),
		}, {
			content: []byte("\n\nb"),
		}, nil, {
			content: []byte("A2"),
		}},
	}, {
		desc:    "with cell formatting",
		content: "3*|DUP\n^|A2\n3*x|B2\n>|C2",
		exp: []*tableCell{{
			content: []byte("DUP\n"),
			format: cellFormat{
				ndupCol: 3,
			},
		}, {
			content: []byte("A2\n3*x"),
			format: cellFormat{
				alignHor: colAlignMiddle,
			},
		}, {
			content: []byte("B2\n"),
		}, {
			content: []byte("C2"),
			format: cellFormat{
				alignHor: colAlignBottom,
			},
		}},
	}}

	for _, c := range cases {
		pt := newTableParser([]byte(c.content))
		test.Assert(t, c.desc, c.exp, pt.cells)
	}
}
