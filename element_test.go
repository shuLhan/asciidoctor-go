// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/math/big"
	"github.com/shuLhan/share/lib/test"
)

func TestAdocNode_parseListDescriptionItem(t *testing.T) {
	cases := []struct {
		line       string
		expLevel   int
		expRawTerm string
		expRaw     string
	}{{
		line:       "CPU::",
		expLevel:   0,
		expRawTerm: "CPU",
	}}

	for _, c := range cases {
		el := &element{}
		el.parseListDescriptionItem([]byte(c.line))

		test.Assert(t, "element.Level", c.expLevel, el.level, true)
		test.Assert(t, "element.rawLabel", c.expRawTerm, el.rawLabel.String(), true)
		test.Assert(t, "element.raw", c.expRaw, string(el.raw), true)
	}
}

func TestAdocNode_postConsumeTable(t *testing.T) {
	cases := []struct {
		desc string
		raw  string
		exp  elementTable
	}{{
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

	for _, c := range cases {
		el := &element{
			raw: []byte(c.raw),
		}
		got := el.postConsumeTable()
		test.Assert(t, c.desc, c.exp, *got, false)
	}
}
