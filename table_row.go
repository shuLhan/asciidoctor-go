// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "bytes"

type tableRow struct {
	cells []*tableCell
	ncell int
}

func (row *tableRow) add(cell *tableCell) {
	row.cells = append(row.cells, cell)
	row.ncell++
	for x := 1; x < cell.format.ndupCol; x++ {
		row.cells = append(row.cells, cell)
		row.ncell++
	}
	if cell.format.nspanCol > 0 {
		row.ncell += (cell.format.nspanCol - 1)
	}
}

func (row *tableRow) String() string {
	var buf bytes.Buffer
	for _, cell := range row.cells {
		buf.WriteByte('|')
		buf.Write(cell.content)
	}
	return buf.String()
}
