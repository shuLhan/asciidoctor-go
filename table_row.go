// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "bytes"

type tableRow struct {
	cells []*tableCell
	ncell int
}

func (row *tableRow) add(cell *tableCell) {
	row.cells = append(row.cells, cell)
	row.ncell++

	var x int
	for x = 1; x < cell.format.ndupCol; x++ {
		row.cells = append(row.cells, cell)
		row.ncell++
	}
	if cell.format.nspanCol > 0 {
		row.ncell += (cell.format.nspanCol - 1)
	}
}

func (row *tableRow) String() string {
	var (
		buf  bytes.Buffer
		cell *tableCell
	)

	for _, cell = range row.cells {
		buf.WriteByte('|')
		buf.Write(cell.content)
	}

	return buf.String()
}
