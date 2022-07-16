// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/math/big"
	"github.com/shuLhan/share/lib/test"
)

func TestAdocNode_parseListDescriptionItem(t *testing.T) {
	type testCase struct {
		line       string
		expRawTerm string
		expRaw     string
		expLevel   int
	}

	var cases = []testCase{{
		line:       "CPU::",
		expLevel:   0,
		expRawTerm: "CPU",
	}}

	var (
		el *element
		c  testCase
	)

	for _, c = range cases {
		el = &element{}
		el.parseListDescriptionItem([]byte(c.line))

		test.Assert(t, "element.Level", c.expLevel, el.level)
		test.Assert(t, "element.rawLabel", c.expRawTerm, el.rawLabel.String())
		test.Assert(t, "element.raw", c.expRaw, string(el.raw))
	}
}

func TestElement_parseListUnorderedItem(t *testing.T) {
	type testCase struct {
		desc     string
		line     []byte
		expRaw   []byte
		expRoles []string
		expLevel int
	}

	var cases = []testCase{{
		desc:     "With text",
		line:     []byte("* \t a"),
		expRaw:   []byte("a\n"),
		expLevel: 1,
	}, {
		desc:     "With unchecked box, no text",
		line:     []byte("* [ ]"),
		expRaw:   []byte("[ ]\n"),
		expLevel: 1,
	}, {
		desc:     "With unchecked box",
		line:     []byte("* [ ] \t a"),
		expRaw:   []byte("&#10063; a\n"),
		expRoles: []string{classNameChecklist},
		expLevel: 1,
	}, {
		desc:     "With checked box, using 'x'",
		line:     []byte("* [x] \t a"),
		expRaw:   []byte("&#10003; a\n"),
		expRoles: []string{classNameChecklist},
		expLevel: 1,
	}, {
		desc:     "With checked box, using 'X'",
		line:     []byte("* [X] \t a"),
		expRaw:   []byte("&#10003; a\n"),
		expRoles: []string{classNameChecklist},
		expLevel: 1,
	}, {
		desc:     "With checked box, using '*'",
		line:     []byte("* [*] \t a"),
		expRaw:   []byte("&#10003; a\n"),
		expRoles: []string{classNameChecklist},
		expLevel: 1,
	}}

	var (
		c  testCase
		el *element
	)

	for _, c = range cases {
		el = &element{}
		el.raw = el.raw[:0]
		el.parseListUnorderedItem(c.line)

		test.Assert(t, c.desc+" - level", c.expLevel, el.level)
		test.Assert(t, c.desc+" - roles", c.expRoles, el.roles)
		test.Assert(t, c.desc+" - raw", c.expRaw, el.raw)
	}
}

func TestAdocNode_postConsumeTable(t *testing.T) {
	type testCase struct {
		desc string
		raw  string
		exp  elementTable
	}

	var cases = []testCase{{
		desc: "single row, multiple lines",
		raw:  "|A\n|B",
		exp: elementTable{
			ncols: 2,
			classes: attributeClass{
				classNameTableblock,
				classNameFrameAll,
				classNameGridAll,
				classNameStretch,
			},
			rows: []*tableRow{{
				cells: []*tableCell{{
					content: []byte("A\n"),
				}, {
					content: []byte("B"),
				}},
				ncell: 2,
			}},
			formats: []*columnFormat{{
				width: big.NewRat(50),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}, {
				width: big.NewRat(50),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}},
			hasHeader: false,
		},
	}, {
		desc: "with header",
		raw:  "A1|B1\n\n|A2\n|B2",
		exp: elementTable{
			ncols: 2,
			classes: attributeClass{
				classNameTableblock,
				classNameFrameAll,
				classNameGridAll,
				classNameStretch,
			},
			rows: []*tableRow{{
				cells: []*tableCell{{
					content: []byte("A1"),
				}, {
					content: []byte("B1"),
				}},
				ncell: 2,
			}, {
				cells: []*tableCell{{
					content: []byte("A2\n"),
				}, {
					content: []byte("B2"),
				}},
				ncell: 2,
			}},
			formats: []*columnFormat{{
				width: big.NewRat(50),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}, {
				width: big.NewRat(50),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}},
			hasHeader: true,
		},
	}, {
		desc: "with multiple rows",
		raw:  "A|B|\n\n|r1c1\n|r1c2|\n\n|r2c1 | r2c2",
		exp: elementTable{
			ncols: 3,
			classes: attributeClass{
				classNameTableblock,
				classNameFrameAll,
				classNameGridAll,
				classNameStretch,
			},
			rows: []*tableRow{{
				cells: []*tableCell{{
					content: []byte("A"),
				}, {
					content: []byte("B"),
				}, {
					content: []byte(""),
				}},
				ncell: 3,
			}, {
				cells: []*tableCell{{
					content: []byte("r1c1\n"),
				}, {
					content: []byte("r1c2"),
				}, {
					content: []byte(""),
				}},
				ncell: 3,
			}},
			formats: []*columnFormat{{
				width: big.NewRat("33.3333"),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}, {
				width: big.NewRat("33.3333"),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}, {
				width: big.NewRat("33.3334"),
				classes: []string{
					classNameTableBlock,
					classNameHalignLeft,
					classNameValignTop,
				},
			}},
			hasHeader: true,
		},
	}}

	var (
		c   testCase
		el  *element
		got *elementTable
	)

	for _, c = range cases {
		el = &element{
			raw: []byte(c.raw),
		}
		got = el.postConsumeTable()
		test.Assert(t, c.desc, c.exp, *got)
	}
}
