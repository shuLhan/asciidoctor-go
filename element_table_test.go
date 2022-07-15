// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/math/big"
	"github.com/shuLhan/share/lib/test"
)

func TestParseAttrCols(t *testing.T) {
	cases := []struct {
		val     string
		formats []*columnFormat
		ncols   int
	}{{
		val:   "3*",
		ncols: 3,
		formats: []*columnFormat{
			newColumnFormat(),
			newColumnFormat(),
			newColumnFormat(),
		},
	}, {
		val:   "3*^",
		ncols: 3,
		formats: []*columnFormat{{
			alignHor: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignHor: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignHor: colAlignMiddle,
			width:    big.NewRat(1),
		}},
	}, {
		val:   "2*,^",
		ncols: 2,
		formats: []*columnFormat{{
			width: big.NewRat(1),
		}, {
			alignHor: colAlignMiddle,
			width:    big.NewRat(1),
		}},
	}, {
		val:   "<,^,>",
		ncols: 3,
		formats: []*columnFormat{{
			alignHor: colAlignTop,
			width:    big.NewRat(1),
		}, {
			alignHor: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignHor: colAlignBottom,
			width:    big.NewRat(1),
		}},
	}, {
		val:   "3*.^",
		ncols: 3,
		formats: []*columnFormat{{
			alignVer: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignVer: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignVer: colAlignMiddle,
			width:    big.NewRat(1),
		}},
	}, {
		val:   "2*,.>",
		ncols: 2,
		formats: []*columnFormat{{
			width: big.NewRat(1),
		}, {
			alignVer: colAlignBottom,
			width:    big.NewRat(1),
		}},
	}, {
		val:   ".<,.^,.>",
		ncols: 3,
		formats: []*columnFormat{{
			alignVer: colAlignTop,
			width:    big.NewRat(1),
		}, {
			alignVer: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignVer: colAlignBottom,
			width:    big.NewRat(1),
		}},
	}, {
		val:   ".<,.^,^.>",
		ncols: 3,
		formats: []*columnFormat{{
			alignVer: colAlignTop,
			width:    big.NewRat(1),
		}, {
			alignVer: colAlignMiddle,
			width:    big.NewRat(1),
		}, {
			alignHor: colAlignMiddle,
			alignVer: colAlignBottom,
			width:    big.NewRat(1),
		}},
	}, {
		val:   "1,2,6",
		ncols: 3,
		formats: []*columnFormat{{
			width: big.NewRat(1),
		}, {
			width: big.NewRat(2),
		}, {
			width: big.NewRat(6),
		}},
	}, {
		val:   "50,20,30",
		ncols: 3,
		formats: []*columnFormat{{
			width: big.NewRat(50),
		}, {
			width: big.NewRat(20),
		}, {
			width: big.NewRat(30),
		}},
	}, {
		val:   ".<2,.^5,^.>3",
		ncols: 3,
		formats: []*columnFormat{{
			alignVer: colAlignTop,
			width:    big.NewRat(2),
		}, {
			alignVer: colAlignMiddle,
			width:    big.NewRat(5),
		}, {
			alignHor: colAlignMiddle,
			alignVer: colAlignBottom,
			width:    big.NewRat(3),
		}},
	}, {
		val:   "h,m,s,e",
		ncols: 4,
		formats: []*columnFormat{{
			style: colStyleHeader,
			width: big.NewRat(1),
		}, {
			style: colStyleMonospaced,
			width: big.NewRat(1),
		}, {
			style: colStyleStrong,
			width: big.NewRat(1),
		}, {
			style: colStyleEmphasis,
			width: big.NewRat(1),
		}},
	}}

	for _, c := range cases {
		ncols, formats := parseAttrCols(c.val)
		test.Assert(t, "ncols", c.ncols, ncols)
		test.Assert(t, c.val, c.formats, formats)
	}
}

func TestParseToRawRows(t *testing.T) {
	cases := []struct {
		desc string
		raw  string
		exp  [][]byte
	}{{
		desc: "empty content",
		raw:  ``,
		exp:  make([][]byte, 1),
	}, {
		desc: "empty content with no header",
		raw: `
`,
		exp: make([][]byte, 2),
	}, {
		desc: "header only",
		raw:  `A | B`,
		exp: [][]byte{
			[]byte("A | B"),
		},
	}, {
		desc: "without header",
		raw: `
| A`,
		exp: [][]byte{
			nil,
			[]byte("| A"),
		},
	}, {
		desc: "header and one row",
		raw: `A
| B

| C`,
		exp: [][]byte{
			[]byte("A"),
			[]byte("| B"),
			nil,
			[]byte("| C"),
		},
	}, {
		desc: "with no header, consecutive rows",
		raw: `
| A | B
| C | D`,
		exp: [][]byte{
			nil,
			[]byte("| A | B"),
			[]byte("| C | D"),
		},
	}}

	for _, c := range cases {
		got := parseToRawRows([]byte(c.raw))
		test.Assert(t, c.desc, c.exp, got)
	}
}
