// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
	l := len(tc.content)
	if l == 0 {
		return false
	}
	return tc.content[l-1] == '\n'
}
