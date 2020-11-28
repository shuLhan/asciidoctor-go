// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"strings"

	"github.com/shuLhan/share/lib/math/big"
)

type adocTable struct {
	ncols     int
	rows      []*tableRow
	formats   []*columnFormat
	hasHeader bool
}

func newTable(attrCols string, content []byte) (table *adocTable) {
	var (
		row *tableRow
	)

	table = &adocTable{}
	table.ncols, table.formats = parseAttrCols(attrCols)

	pt := newParserTable(content)

	if table.ncols == 0 {
		row = pt.firstRow()
		if row.ncell == 0 {
			return table
		}
		table.ncols = row.ncell
	} else {
		row = pt.row(table.ncols)
	}
	if pt.nrow == 1 && !row.cells[0].endWithLF() {
		table.hasHeader = true
	}
	for row.ncell == table.ncols {
		table.rows = append(table.rows, row)
		row = pt.row(table.ncols)
	}

	if len(table.formats) == 0 {
		for x := 0; x < table.ncols; x++ {
			table.formats = append(table.formats, newColumnFormat())
		}
	}

	table.recalculateWidth()
	table.initializeFormats()

	return table
}

func (table *adocTable) initializeFormats() {
	for _, format := range table.formats {
		classes := []string{classNameTableBlock}

		switch format.alignHor {
		case colAlignTop:
			classes = append(classes, classNameHalignLeft)
		case colAlignMiddle:
			classes = append(classes, classNameHalignCenter)
		case colAlignBottom:
			classes = append(classes, classNameHalignRight)
		}
		switch format.alignVer {
		case colAlignTop:
			classes = append(classes, classNameValignTop)
		case colAlignMiddle:
			classes = append(classes, classNameValignMiddle)
		case colAlignBottom:
			classes = append(classes, classNameValignBottom)
		}
		format.classes = classes
	}
}

func (table *adocTable) recalculateWidth() {
	var (
		totalWidth = big.NewRat(0)
	)
	for _, format := range table.formats {
		totalWidth.Add(format.width)
	}
	for _, format := range table.formats {
		format.width = big.QuoRat(format.width, totalWidth).Mul(100)
	}
}

//
// parseAttrCols parse the value of attribute "cols=".
//
//	ATTR_COLS = (NCOLS ("*")) / COL_ATTR  ("," COL_ATTR)
//
//	NCOLS     = 1*DIGITS
//
//	COL_ATTR  = (".") ( "<" / "^" / ">" ) ( COL_WIDTH ) (COL_STYLE)
//
//	COL_WIDTH = DIGITS / (2DIGITs)
//
//	COL_STYLE = "a" / "e" / "h" / "l" / "m" / "d" / "s" / "v"
//
func parseAttrCols(val string) (ncols int, formats []*columnFormat) {
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return 0, nil
	}

	var hasDefault bool

	rawFormat := strings.Split(val, ",")

	n, format := parseColumnFormat(rawFormat[0])
	if format.isDefault {
		hasDefault = true
		ncols = n
		format.isDefault = false
		for x := 0; x < n; x++ {
			f := newColumnFormat()
			f.merge(format)
			formats = append(formats, f)
		}
	} else {
		ncols = len(rawFormat)
		formats = append(formats, format)
	}

	idxFromLast := ncols - (len(rawFormat) - 1)
	for x := 1; x < len(rawFormat); x++ {
		_, format = parseColumnFormat(rawFormat[x])
		if hasDefault {
			formats[idxFromLast].merge(format)
			idxFromLast++
		} else {
			formats = append(formats, format)
		}
	}

	return ncols, formats
}

//
// parseToRawRows convert raw table content into multiple raw rows.
//
func parseToRawRows(raw []byte) (rows [][]byte) {
	lines := bytes.Split(raw, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		rows = append(rows, line)
	}
	return rows
}
