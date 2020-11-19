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
		node := &adocNode{}
		node.parseListDescriptionItem(c.line)

		test.Assert(t, "adocNode.Level", c.expLevel, node.level, true)
		test.Assert(t, "adocNode.rawLabel", c.expRawTerm, node.rawLabel.String(), true)
		test.Assert(t, "adocNode.raw", c.expRaw, string(node.raw), true)
	}
}

func TestAdocNode_postConsumeTable(t *testing.T) {
	cases := []struct {
		desc string
		raw  string
		exp  adocTable
	}{{
		desc: "without header",
		raw:  "|A\n|B",
		exp: adocTable{
			ncols: 2,
			rows:  []tableRow{{"A", "B"}},
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
		},
	}, {
		desc: "with header",
		raw:  "A|B\n\n|r1-c1\n|r1-c2",
		exp: adocTable{
			ncols:  2,
			header: tableRow{"A", "B"},
			rows:   []tableRow{{"r1-c1", "r1-c2"}},
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
		},
	}, {
		desc: "with multiple rows",
		raw:  "A|B|\n\n|r1c1\n|r1c2|\n\n|r2c1 | r2c2",
		exp: adocTable{
			ncols:  2,
			header: tableRow{"A", "B"},
			rows: []tableRow{
				{"r1c1", "r1c2"},
				{"r2c1", "r2c2"},
			},
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
		},
	}}

	for _, c := range cases {
		node := &adocNode{
			raw: []byte(c.raw),
		}
		got := node.postConsumeTable()
		test.Assert(t, c.desc, c.exp, *got, false)
	}
}
