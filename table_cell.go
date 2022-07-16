// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

type tableCell struct {
	content []byte
	format  cellFormat
}

func (tc *tableCell) writeString(s string) {
	tc.content = append(tc.content, []byte(s)...)
}

func (tc *tableCell) writeByte(b byte) {
	tc.content = append(tc.content, b)
}

func (tc *tableCell) endWithLF() bool {
	var l int = len(tc.content)
	if l == 0 {
		return false
	}
	return tc.content[l-1] == '\n'
}
