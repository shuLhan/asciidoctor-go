// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/math/big"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
)

type elementTable struct {
	styles map[string]string

	rows    []*tableRow
	formats []*columnFormat

	classes attributeClass
	ncols   int

	hasHeader bool
	hasFooter bool
}

func newTable(ea *elementAttribute, content []byte) (table *elementTable) {
	var (
		row       *tableRow
		pt        *tableParser
		attrValue string
	)

	table = &elementTable{
		classes: attributeClass{
			classNameTableblock,
			classNameFrameAll,
			classNameGridAll,
		},
		styles: make(map[string]string),
	}

	attrValue = ea.Attrs[attrNameCols]
	if len(attrValue) > 0 {
		table.ncols, table.formats = parseAttrCols(attrValue)
	}

	table.parseOptions(ea.options)

	pt = newTableParser(content)

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
		if !libstrings.IsContain(ea.options, attrValueNoHeader) {
			if pt.x < len(pt.cells) && pt.cells[pt.x] == nil {
				table.hasHeader = true
			}
		}
	}

	for row.ncell == table.ncols {
		table.rows = append(table.rows, row)
		row = pt.row(table.ncols)
	}
	if len(table.rows) == 1 {
		if !libstrings.IsContain(ea.options, attrValueHeader) {
			table.hasHeader = false
		}
	}

	if len(table.formats) == 0 {
		var x int
		for x = 0; x < table.ncols; x++ {
			table.formats = append(table.formats, newColumnFormat())
		}
	}

	table.recalculateWidth()
	table.initializeFormats()
	table.initializeClassAndStyles(ea)

	return table
}

func (table *elementTable) initializeFormats() {
	var (
		format *columnFormat
	)

	for _, format = range table.formats {
		var classes = []string{classNameTableBlock}

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

func (table *elementTable) initializeClassAndStyles(ea *elementAttribute) {
	var (
		k         string
		v         string
		withWidth bool
	)
	for k, v = range ea.Attrs {
		switch k {
		case attrNameWidth:
			if len(v) == 0 {
				continue
			}
			if v[len(v)-1] != '%' {
				v += `%`
			}
			table.styles[k] = v
			withWidth = true
		case attrNameFrame:
			switch v {
			case attrValueTopbot:
				table.classes.replace(classNameFrameAll,
					classNameFrameEnds)
			case attrValueSides:
				table.classes.replace(classNameFrameAll,
					classNameFrameSides)
			case attrValueNone:
				table.classes.replace(classNameFrameAll,
					classNameFrameNone)
			}
		case attrNameGrid:
			switch v {
			case attrValueCols:
				table.classes.replace(classNameGridAll,
					classNameGridCols)
			case attrValueNone:
				table.classes.replace(classNameGridAll,
					classNameGridNone)
			case attrValueRows:
				table.classes.replace(classNameGridAll,
					classNameGridRows)
			}
		case attrNameStripes:
			switch v {
			case attrValueAll:
				table.classes.add(classNameStripesAll)
			case attrValueEven:
				table.classes.add(classNameStripesEven)
			case attrValueHover:
				table.classes.add(classNameStripesHover)
			case attrValueOdd:
				table.classes.add(classNameStripesOdd)
			}
		}
	}
	for _, k = range ea.options {
		if k == optNameAutowidth {
			withWidth = true
			table.classes.add(classNameFitContent)

			var f *columnFormat
			for _, f = range table.formats {
				f.width = nil
			}
		}
	}
	for _, k = range ea.roles {
		table.classes.add(k)
	}
	if !withWidth {
		table.classes.add(classNameStretch)
	}
}

func (table *elementTable) parseOptions(opts []string) {
	if opts == nil {
		return
	}
	var key string
	for _, key = range opts {
		switch key {
		case attrValueHeader:
			table.hasHeader = true
		case attrValueFooter:
			table.hasFooter = true
		}
	}
}

func (table *elementTable) recalculateWidth() {
	var (
		totalWidth = big.NewRat(0)
		lastWidth  = big.NewRat(100)

		format       *columnFormat
		x            int
		hasAutowidth bool
	)
	for _, format = range table.formats {
		if format.isAutowidth {
			hasAutowidth = true
			format.width = nil
		} else {
			totalWidth.Add(format.width)
		}
	}
	for x, format = range table.formats {
		if hasAutowidth {
			continue
		}
		if x == len(table.formats)-1 {
			format.width = lastWidth
		} else {
			format.width = big.QuoRat(format.width, totalWidth).Mul(100)
			lastWidth.Sub(format.width)
		}
	}
}

func (table *elementTable) htmlStyle() string {
	var (
		buf bytes.Buffer
		k   string
		v   string
	)

	for k, v = range table.styles {
		fmt.Fprintf(&buf, `%s: %s;`, k, v)
	}
	return buf.String()
}

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
func parseAttrCols(val string) (ncols int, formats []*columnFormat) {
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return 0, nil
	}

	var (
		rawFormat = strings.Split(val, `,`)

		format      *columnFormat
		n           int
		x           int
		idxFromLast int
		hasDefault  bool
	)

	n, format = parseColumnFormat(rawFormat[0])
	if format.isDefault {
		hasDefault = true
		ncols = n
		format.isDefault = false
		for x = 0; x < n; x++ {
			var f = newColumnFormat()
			f.merge(format)
			formats = append(formats, f)
		}
	} else {
		ncols = len(rawFormat)
		formats = append(formats, format)
	}

	idxFromLast = ncols - (len(rawFormat) - 1)
	for x = 1; x < len(rawFormat); x++ {
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

// parseToRawRows convert raw table content into multiple raw rows.
func parseToRawRows(raw []byte) (rows [][]byte) {
	var (
		lines = bytes.Split(raw, []byte{'\n'})
		line  []byte
	)

	for _, line = range lines {
		line = bytes.TrimSpace(line)
		rows = append(rows, line)
	}
	return rows
}
