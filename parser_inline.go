// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// parserInline is the one the responsible to parse text that contains inline
// markup (bold, italic, etc.) into tree.
//
type parserInline struct {
	container      *adocNode
	current        *adocNode
	content        []byte
	x              int
	state          *parserInlineState
	prev, c, nextc byte
}

func newParserInline(content []byte) (pi *parserInline) {
	pi = &parserInline{
		container: &adocNode{
			kind: nodeKindText,
		},
		content: bytes.TrimRight(content, "\n"),
		state:   &parserInlineState{},
	}
	pi.current = pi.container

	return pi
}

func (pi *parserInline) next() {
	pi.c = pi.content[pi.x]
	if pi.x+1 == len(pi.content) {
		pi.nextc = 0
	} else {
		pi.nextc = pi.content[pi.x+1]
	}
}

func (pi *parserInline) do() {
	for pi.x < len(pi.content) {
		pi.next()

		if pi.c == '+' {
			if pi.nextc == '+' {
				pi.parsePassthroughDouble()
			} else {
				pi.parsePassthrough()
			}
			continue
		}
		if pi.c == '"' {
			if pi.nextc == '`' {
				ok := pi.parseQuoteBegin([]byte("`\""),
					nodeKindSymbolQuoteDoubleBegin)
				if ok {
					continue
				}
			}
			pi.current.WriteByte('"')
			pi.x++
			pi.prev = pi.c
			continue
		}
		if pi.c == '\'' {
			if pi.nextc == '`' {
				ok := pi.parseQuoteBegin([]byte("`'"),
					nodeKindSymbolQuoteSingleBegin)
				if ok {
					continue
				}
			}
			pi.current.WriteByte('\'')
			pi.x++
			pi.prev = pi.c
			continue
		}
		if pi.c == '*' {
			if pi.nextc == '*' {
				if pi.parseFormatUnconstrained(
					[]byte("**"),
					nodeKindUnconstrainedBold,
					nodeKindTextBold,
					styleTextBold) {
					continue
				}
			}
			pi.parseFormat(nodeKindTextBold, styleTextBold)
			continue
		}
		if pi.c == '_' {
			if pi.nextc == '_' {
				if pi.parseFormatUnconstrained(
					[]byte("__"),
					nodeKindUnconstrainedItalic,
					nodeKindTextItalic,
					styleTextItalic) {
					continue
				}
			}
			pi.parseFormat(nodeKindTextItalic, styleTextItalic)
			continue
		}
		if pi.c == '`' {
			var ok bool
			if pi.nextc == '`' {
				ok = pi.parseFormatUnconstrained(
					[]byte("``"),
					nodeKindUnconstrainedMono,
					nodeKindTextMono,
					styleTextMono)
				if ok {
					continue
				}
			}
			if pi.nextc == '"' {
				if pi.parseQuoteEnd([]byte("`\""),
					nodeKindSymbolQuoteDoubleEnd) {
					continue
				}
			}
			if pi.nextc == '\'' {
				if pi.parseQuoteEnd([]byte("`'"),
					nodeKindSymbolQuoteSingleEnd) {
					continue
				}

				// This is an aposthrope
				pi.current.WriteString(symbolQuoteSingleEnd)
				pi.x += 2
				pi.prev = 0
				continue
			}
			pi.parseFormat(nodeKindTextMono, styleTextMono)
			continue
		}
		pi.current.WriteByte(pi.c)
		pi.x++
		pi.prev = pi.c
	}
}

//
// parseQuoteBegin check if the double quote curve ("`) is valid (does not
// followed by space) and has an end (`")..
//
func (pi *parserInline) parseQuoteBegin(quoteEnd []byte, kind int) bool {
	if pi.x+2 >= len(pi.content) {
		return false
	}
	c := pi.content[pi.x+2]
	if ascii.IsSpace(c) {
		return false
	}
	raw := pi.content[pi.x+2:]
	idx := bytes.LastIndex(raw, quoteEnd)
	if idx < 0 {
		return false
	}
	if ascii.IsSpace(raw[idx-1]) {
		return false
	}
	node := &adocNode{
		kind: kind,
	}
	pi.current.addChild(node)
	pi.current = node
	pi.x += 2
	pi.prev = 0
	return true
}

func (pi *parserInline) parseQuoteEnd(quoteEnd []byte, kind int) bool {
	if ascii.IsSpace(pi.prev) {
		// This is not the end that we looking for.
		return false
	}
	node := &adocNode{
		kind: kind,
	}
	pi.current.addChild(node)
	pi.current = node
	pi.x += 2
	pi.prev = 0
	return true
}

func (pi *parserInline) parseFormat(kind int, style int64) bool {
	// Do we have the begin format?
	if isEndFormat(pi.nextc) {
		if pi.state.has(kind) {
			pi.terminate(kind, style)
			pi.prev = 0
			pi.x++
			return true
		}
	}

	// Do we have the end format?
	raw := pi.content[pi.x+1:]
	idx := bytes.LastIndexByte(raw, pi.c)
	if idx > 0 {
		var end byte
		if idx+1 < len(raw) {
			end = raw[idx+1]
		}
		if isEndFormat(end) {
			if isBeginFormat(pi.prev, pi.nextc) {
				if !pi.state.has(kind) {
					node := &adocNode{
						kind: kind,
					}
					pi.current.addChild(node)
					pi.state.push(kind)
					pi.current = node
					pi.prev = 0
					pi.x++
					return true
				}
			}
		}
	}

	// No 'c' termination found.
	pi.current.WriteByte(pi.c)
	pi.x++
	pi.prev = pi.c
	return false
}

func (pi *parserInline) parseFormatUnconstrained(
	terms []byte,
	kindUnconstrained int,
	kind int,
	style int64,
) bool {
	// Have we parsed the unconstrained format before?
	if pi.state.has(kindUnconstrained) {
		pi.terminate(kindUnconstrained, style)
		pi.prev = 0
		pi.x += 2
		return true
	}
	// Have we parsed single format before?
	if pi.state.has(kind) {
		pi.current.WriteByte(pi.c)
		pi.terminate(kind, style)
		pi.prev = 0
		pi.x += 2
		return true
	}

	// Do we have the end format?
	raw := pi.content[pi.x+2:]
	if bytes.Contains(raw, terms) {
		node := &adocNode{
			kind: kindUnconstrained,
		}
		pi.current.addChild(node)
		pi.state.push(kindUnconstrained)
		pi.current = node
		pi.prev = 0
		pi.x += 2
		return true
	}

	return false
}

func (pi *parserInline) parsePassthroughDouble() bool {
	raw := pi.content[pi.x+2:]

	// Check if we have "++" at the end.
	idx := bytes.Index(raw, []byte("++"))
	if idx >= 0 {
		node := &adocNode{
			kind: nodeKindPassthroughDouble,
			raw:  libbytes.Copy(raw[:idx]),
		}
		pi.current.addChild(node)
		pi.current = node
		pi.x += idx + 4
		pi.prev = 0
		return true
	}

	// Check if we have single '+'...
	idx = bytes.IndexByte(raw, pi.c)
	if idx >= 0 {
		raw = pi.content[pi.x+1:]
		idx++
		node := &adocNode{
			kind: nodeKindPassthrough,
			raw:  libbytes.Copy(raw[:idx]),
		}
		pi.current.addChild(node)
		pi.current = node
		pi.x += idx + 2
		pi.prev = 0
		return true
	}

	// No '++' or '+' found as termination.
	pi.current.WriteString("++")
	pi.x += 2
	return false
}

func (pi *parserInline) parsePassthrough() bool {
	raw := pi.content[pi.x+1:]
	idx := bytes.IndexByte(raw, pi.c)
	if idx >= 0 {
		node := &adocNode{
			kind: nodeKindPassthrough,
			raw:  libbytes.Copy(raw[:idx]),
		}
		pi.current.addChild(node)
		pi.current = node
		pi.x += idx + 2
		pi.prev = 0
		return true
	}

	// No '+' found as termination.
	pi.current.WriteByte(pi.c)
	pi.prev = pi.c
	pi.x++
	return false
}

func (pi *parserInline) terminate(kind int, style int64) {
	node := pi.current
	stateTmp := &parserInlineState{}
	for node.parent != nil {
		if node.kind == kind {
			pi.state.pop()
			node.style |= style
			break
		}
		if node.kind == nodeKindTextBold && node.style == 0 {
			node.style = styleTextBold
			stateTmp.push(pi.state.pop())
		}
		if node.kind == nodeKindTextItalic && node.style == 0 {
			node.style = styleTextItalic
			stateTmp.push(pi.state.pop())
		}
		if node.kind == nodeKindTextMono && node.style == 0 {
			node.style = styleTextMono
			stateTmp.push(pi.state.pop())
		}
		node = node.parent
	}
	if node.parent != nil {
		node = node.parent
	}
	for k := stateTmp.pop(); k != 0; k = stateTmp.pop() {
		child := &adocNode{
			kind: k,
		}
		node.addChild(child)
		node = child
		pi.state.push(k)
	}
	child := &adocNode{
		kind: nodeKindText,
	}
	node.addChild(child)
	pi.current = child
}

func isBeginFormat(prev, next byte) bool {
	if prev == ':' || prev == ';' || ascii.IsAlnum(prev) {
		return false
	}
	if ascii.IsSpace(next) || next == 0 {
		return false
	}
	return true
}

func isEndFormat(next byte) bool {
	if ascii.IsAlnum(next) {
		return false
	}
	if next == 0 || next == ':' || next == '*' || next == '.' || next == '_' || ascii.IsSpace(next) {
		return true
	}
	return false
}
