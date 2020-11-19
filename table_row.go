// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
)

type tableRow []string

//
// parseTableRow parse a line into one or more columns.
// Each column is separated by '|'.
//
func parseTableRow(line []byte) (row tableRow) {
	if len(line) == 0 {
		return nil
	}
	if line[0] == '|' {
		line = line[1:]
	}

	// Chop line into multiple columns.
	for len(line) > 0 {
		col, idx := indexByteUnescape(line, '|')
		col = bytes.TrimSpace(col)
		if len(col) > 0 {
			row = append(row, string(col))
		}
		if idx < 0 {
			line = bytes.TrimSpace(line)
			if len(line) > 0 {
				row = append(row, string(line))
			}
			break
		}
		line = line[idx+1:]
	}
	return row
}

func parseTableRows(ncols int, lines [][]byte) (row tableRow, rest [][]byte) {
	x := 0
	for x < len(lines) {
		tmp := parseTableRow(lines[x])
		if ncols == 0 {
			if tmp == nil {
				// We got row separator.
				return row, lines[x+1:]
			}
		} else {
			if len(row) == ncols {
				return row, lines[x+1:]
			}
		}
		row = append(row, tmp...)
		x++
	}
	return row, rest
}
