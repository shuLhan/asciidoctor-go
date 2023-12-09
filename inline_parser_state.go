// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

type inlineParserState struct {
	stack []int
}

func (pis *inlineParserState) push(c int) {
	pis.stack = append(pis.stack, c)
}

func (pis *inlineParserState) pop() (c int) {
	var size = len(pis.stack)
	if size > 0 {
		c = pis.stack[size-1]
		pis.stack = pis.stack[:size-1]
	}
	return c
}

func (pis *inlineParserState) has(c int) bool {
	var r int

	for _, r = range pis.stack {
		if r == c {
			return true
		}
	}
	return false
}
