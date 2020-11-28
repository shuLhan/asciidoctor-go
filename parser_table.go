// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"strings"

	"github.com/shuLhan/share/lib/parser"
)

type parserTable struct {
	p     *parser.Parser
	cells []*tableCell
	nrow  int
	x     int
}

func newParserTable(content []byte) (pt *parserTable) {
	pt = &parserTable{
		p: parser.New(string(content), "|\n"),
	}
	pt.toCells()
	return pt
}

//
// toCells parse the raw table content into cells.
//
func (pt *parserTable) toCells() {
	var (
		token, c  = pt.p.TokenEscaped('\\')
		tokenTrim = strings.TrimSpace(token)
		l         = len(tokenTrim)
		cell      = &tableCell{}
	)

	// Parse the first cell with three possibilities,
	// Case 1: cell without '|', for example "...\n|..."
	// Case 2: empty line, for example "\n\n|..."
	// Case 3: cell with formatting "3*|..."
	for c == '\n' {
		if l > 0 {
			// Case 1.
			cell.writeString(token)
			cell.writeByte('\n')
		} else {
			// Case 2.
			pt.cells = append(pt.cells, nil)
		}
		token, c = pt.p.TokenEscaped('\\')
		tokenTrim = strings.TrimSpace(token)
		l = len(tokenTrim)
	}
	if c == 0 {
		if l > 0 {
			cell.writeString(token)
			pt.addCell(cell)
		}
		return
	}
	if c == '|' {
		cf := parseCellFormat(token)
		if cf == nil {
			if l > 0 {
				// Case 1.
				cell.writeString(token)
				pt.cells = append(pt.cells, cell)
				cell = &tableCell{}
			}
		} else {
			// Case 3.
			cell.format = *cf
		}
	}

	token, c = pt.p.TokenEscaped('\\')
	tokenTrim = strings.TrimSpace(token)
	l = len(tokenTrim)
	for {
		if c == '\n' {
			if l > 0 {
				cell.writeString(token)
			}
			cell.writeByte('\n')
		} else if c == '|' {
			cf := parseCellFormat(token)
			if cf == nil {
				cell.writeString(token)
				pt.addCell(cell)
				cell = &tableCell{}
			} else {
				pt.addCell(cell)
				cell = &tableCell{
					format: *cf,
				}
			}
		} else {
			cell.writeString(token)
			pt.addCell(cell)
			break
		}
		token, c = pt.p.TokenEscaped('\\')
		tokenTrim = strings.TrimSpace(token)
		l = len(tokenTrim)
	}
}

func (pt *parserTable) addCell(cell *tableCell) {
	var (
		emptyLine        = []byte("\n\n")
		endWithEmptyLine = bytes.HasSuffix(cell.content, emptyLine)
	)
	if endWithEmptyLine {
		cell.content = bytes.TrimSuffix(cell.content, emptyLine)
	}

	pt.cells = append(pt.cells, cell)
	if endWithEmptyLine {
		pt.cells = append(pt.cells, nil)
	}
}

//
// firstRow get the first row of the table to get the number of columns.
//
func (pt *parserTable) firstRow() (row *tableRow) {
	row = &tableRow{}

	// Skip empty lines..
	for ; pt.x < len(pt.cells) && pt.cells[pt.x] == nil; pt.x++ {
		pt.nrow++
	}
	for ; pt.x < len(pt.cells); pt.x++ {
		if pt.cells[pt.x] == nil {
			break
		}
		row.add(pt.cells[pt.x])
	}
	pt.nrow++
	return row
}

//
// row get n number of cells as row, skip any nil cell if exist on list.
//
func (pt *parserTable) row(ncols int) (row *tableRow) {
	row = &tableRow{}
	for ; pt.x < len(pt.cells); pt.x++ {
		if pt.cells[pt.x] == nil {
			continue
		}
		row.add(pt.cells[pt.x])
		if row.ncell == ncols {
			pt.x++
			break
		}
	}
	pt.nrow++
	return row
}
