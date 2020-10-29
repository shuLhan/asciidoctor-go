// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
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
	isEscaped      bool
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

func (pi *parserInline) do() {
	for pi.x < len(pi.content) {
		pi.c = pi.content[pi.x]
		if pi.x+1 == len(pi.content) {
			pi.nextc = 0
		} else {
			pi.nextc = pi.content[pi.x+1]
		}

		if pi.c == '\\' {
			if pi.isEscaped {
				pi.escape()
				pi.prev = 0
				continue
			}
			pi.isEscaped = true
			pi.x++
			pi.prev = pi.c
			continue
		}
		if pi.c == '+' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '+' {
				if pi.x+2 < len(pi.content) && pi.content[pi.x+2] == '+' {
					if pi.parsePassthroughTriple() {
						continue
					}
				}
				if pi.parsePassthroughDouble() {
					continue
				}
			}
			if pi.parsePassthrough() {
				continue
			}
		} else if pi.c == ':' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.parseMacro() {
				continue
			}
		} else if pi.c == '~' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.parseSubscript() {
				continue
			}
		} else if pi.c == '^' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.parseSuperscript() {
				continue
			}
		} else if pi.c == '"' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '`' {
				ok := pi.parseQuoteBegin([]byte("`\""),
					nodeKindSymbolQuoteDoubleBegin)
				if ok {
					continue
				}
			}
		} else if pi.c == '\'' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '`' {
				ok := pi.parseQuoteBegin([]byte("`'"),
					nodeKindSymbolQuoteSingleBegin)
				if ok {
					continue
				}
			}
		} else if pi.c == '*' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '*' {
				if pi.parseFormatUnconstrained(
					[]byte("**"),
					nodeKindUnconstrainedBold,
					nodeKindTextBold,
					styleTextBold) {
					continue
				}
			}
			if pi.parseFormat(nodeKindTextBold, styleTextBold) {
				continue
			}
		} else if pi.c == '_' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '_' {
				if pi.parseFormatUnconstrained(
					[]byte("__"),
					nodeKindUnconstrainedItalic,
					nodeKindTextItalic,
					styleTextItalic) {
					continue
				}
			}
			if pi.parseFormat(nodeKindTextItalic, styleTextItalic) {
				continue
			}
		} else if pi.c == '`' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '`' {
				if pi.parseFormatUnconstrained(
					[]byte("``"),
					nodeKindUnconstrainedMono,
					nodeKindTextMono,
					styleTextMono) {
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
			if pi.parseFormat(nodeKindTextMono, styleTextMono) {
				continue
			}
		}
		pi.current.WriteByte(pi.c)
		pi.x++
		pi.prev = pi.c
	}
}

func (pi *parserInline) escape() {
	pi.isEscaped = false
	pi.current.WriteByte(pi.c)
	pi.x++
	pi.prev = pi.c
}

func (pi *parserInline) getBackMacroName() (macroName string, lastc byte) {
	raw := pi.current.raw
	start := len(raw) - 1
	for start > 0 {
		if !ascii.IsAlpha(raw[start]) {
			return string(raw[start+1:]), raw[start]
		}
		start--
	}
	return string(raw), 0
}

//
// parseQuoteBegin check if the double quote curve ("`) is valid (does not
// followed by space) and has an end (`").
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
	if ascii.IsSpace(raw[idx-1]) || raw[idx-1] == '\\' {
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
	// Do we have the beginning?
	if isEndFormat(pi.prev, pi.nextc) {
		if pi.state.has(kind) {
			pi.terminate(kind, style)
			pi.prev = 0
			pi.x++
			return true
		}
	}

	if !isBeginFormat(pi.prev, pi.nextc) {
		return false
	}
	if pi.state.has(kind) {
		// if c is begin format but we also have unclosed parent
		return false
	}

	raw := pi.content[pi.x+1:]
	_, idx := indexByteUnescape(raw, pi.c)
	var hasEnd bool
	for idx >= 0 {
		var prevc, nextc byte
		if idx > 0 {
			prevc = raw[idx-1]
		}
		if idx+1 < len(raw) {
			nextc = raw[idx+1]
		}
		if isEndFormat(prevc, nextc) {
			hasEnd = true
			break
		}
		raw = raw[idx+1:]
		_, idx = indexByteUnescape(raw, pi.c)
	}
	if !hasEnd {
		return false
	}

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

func (pi *parserInline) parseInlineImage() *adocNode {
	content := pi.content[pi.x+1:]
	_, idx := indexByteUnescape(pi.content[pi.x+1:], ']')
	if idx < 0 {
		return nil
	}

	lineImage := content[:idx+1]
	nodeImage := &adocNode{
		kind:  nodeKindInlineImage,
		Attrs: make(map[string]string),
	}
	if nodeImage.parseImage(string(lineImage)) {
		pi.x += idx + 2
		pi.prev = 0
		return nodeImage
	}
	return nil
}

func (pi *parserInline) parseMacro() bool {
	name, lastc := pi.getBackMacroName()
	if lastc == '\\' || len(name) == 0 {
		return false
	}

	switch name {
	case "":
		return false
	case macroFTP, macroHTTPS, macroHTTP, macroIRC, macroLink, macroMailto:
		node := pi.parseURL(name)
		if node == nil {
			return false
		}

		pi.current.raw = pi.current.raw[:len(pi.current.raw)-len(name)]

		pi.current.addChild(node)
		node = &adocNode{
			kind: nodeKindText,
		}
		pi.current.addChild(node)
		pi.current = node
		return true
	case macroImage:
		node := pi.parseInlineImage()
		if node == nil {
			return false
		}

		pi.current.raw = pi.current.raw[:len(pi.current.raw)-len(name)]

		pi.current.addChild(node)
		node = &adocNode{
			kind: nodeKindText,
		}
		pi.current.addChild(node)
		pi.current = node
		return true
	}
	return false
}

func (pi *parserInline) parsePassthrough() bool {
	if !isBeginFormat(pi.prev, pi.nextc) {
		return false
	}

	var (
		x             int
		pass          []byte
		prev, c, next byte
		isEsc         bool
		content       = pi.content[pi.x+1:]
	)
	for ; x < len(content); x++ {
		c = content[x]
		if x+1 < len(content) {
			next = content[x+1]
		}
		if c == '\\' {
			if isEsc {
				pass = append(pass, '\\')
				isEsc = false
			} else {
				isEsc = true
			}
			prev = c
			continue
		}
		if c == '+' {
			if isEsc {
				pass = append(pass, '+')
				isEsc = false
				continue
			}
			if isEndFormat(prev, next) {
				break
			}
		}
		pass = append(pass, c)
		prev = c
	}
	if x == len(content) {
		return false
	}

	node := &adocNode{
		kind: nodeKindPassthrough,
		raw:  pass,
	}
	pi.current.addChild(node)
	pi.current = node
	pi.x += x + 2
	pi.prev = 0
	return true
}

func (pi *parserInline) parsePassthroughDouble() bool {
	raw := pi.content[pi.x+2:]

	// Check if we have "++" at the end.
	raw, idx := indexUnescape(raw, []byte("++"))
	if idx >= 0 {
		node := &adocNode{
			kind: nodeKindPassthroughDouble,
			raw:  raw,
		}
		pi.current.addChild(node)
		pi.current = node
		pi.x += idx + 4
		pi.prev = 0
		return true
	}

	return false
}

func (pi *parserInline) parsePassthroughTriple() bool {
	raw := pi.content[pi.x+3:]

	// Check if we have "+++" at the end.
	raw, idx := indexUnescape(raw, []byte("+++"))
	if idx >= 0 {
		node := &adocNode{
			kind: nodeKindPassthroughTriple,
			raw:  raw,
		}
		pi.current.addChild(node)
		pi.current = node
		pi.x += idx + 6
		pi.prev = 0
		return true
	}
	return false
}

func (pi *parserInline) parseSubscript() bool {
	var (
		prev byte
		raw  = pi.content[pi.x+1:]
	)
	for x := 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			node := &adocNode{
				kind: nodeKindTextSubscript,
				raw:  raw[:x],
			}
			pi.current.addChild(node)

			node = &adocNode{
				kind: nodeKindText,
			}
			pi.current.addChild(node)
			pi.current = node

			pi.x += x + 2
			pi.prev = pi.c
			return true
		}
		if ascii.IsSpace(raw[x]) {
			break
		}
		prev = raw[x]
	}
	return false
}

func (pi *parserInline) parseSuperscript() bool {
	var (
		prev byte
		raw  = pi.content[pi.x+1:]
	)
	for x := 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			node := &adocNode{
				kind: nodeKindTextSuperscript,
				raw:  raw[:x],
			}
			pi.current.addChild(node)

			node = &adocNode{
				kind: nodeKindText,
			}
			pi.current.addChild(node)
			pi.current = node

			pi.x += x + 2
			pi.prev = pi.c
			return true
		}
		if ascii.IsSpace(raw[x]) {
			break
		}
		prev = raw[x]
	}
	return false
}

//
// parseURL parser the URL, an optional text, optional attribute for target,
// and optional role.
//
// The current state of p.x is equal to ":".
//
func (pi *parserInline) parseURL(scheme string) (node *adocNode) {
	var (
		x   int
		c   byte
		uri []byte
	)
	if scheme != macroLink {
		uri = []byte(scheme)
		uri = append(uri, ':')
	}

	node = &adocNode{
		kind:  nodeKindURL,
		Attrs: make(map[string]string),
	}

	content := pi.content[pi.x+1:]
	for ; x < len(content); x++ {
		c = content[x]
		if c == '[' || ascii.IsSpace(c) {
			break
		}
		uri = append(uri, c)
	}
	if c != '[' {
		if scheme == macroHTTP || scheme == macroHTTPS {
			node.classes = append(node.classes, "bare")
		}
		if !ascii.IsSpace(c) {
			uri = uri[:len(uri)-1]
			pi.prev = 0
			pi.x += x
		} else {
			pi.x += x + 1
			pi.prev = c
		}
		node.raw = uri
		node.Attrs[attrNameHref] = string(uri)
		return node
	}
	_, idx := indexByteUnescape(content[x:], ']')
	if idx < 0 {
		return nil
	}

	node.Attrs[attrNameHref] = string(uri)
	pi.x += x + idx + 2
	pi.prev = 0

	attr := string(content[x : x+idx+1])
	attrs := parseBlockAttribute(attr)
	if len(attrs) == 0 {
		// empty "[]"
		node.raw = uri
		return node
	}
	if len(attrs) >= 1 {
		text := attrs[0]
		if text[len(text)-1] == '^' {
			node.Attrs[attrNameTarget] = attrValueBlank
			text = strings.TrimRight(text, "^")
			node.Attrs[attrNameRel] = attrValueNoopener
		}
		child := parseInlineMarkup([]byte(text))
		node.addChild(child)
	}
	if len(attrs) >= 2 {
		attrTarget := strings.Split(attrs[1], "=")
		if len(attrTarget) == 2 {
			switch attrTarget[0] {
			case attrNameWindow:
				node.Attrs[attrNameTarget] = attrTarget[1]
				if attrTarget[1] == attrValueBlank {
					node.Attrs[attrNameRel] = attrValueNoopener
				}
			case attrNameRole:
				classes := strings.Split(attrTarget[1], ",")
				node.classes = append(node.classes, classes...)
			}
		}
	}
	if len(attrs) >= 3 {
		attrRole := strings.Split(attrs[2], "=")
		if len(attrRole) == 2 {
			switch attrRole[0] {
			case attrNameRole:
				classes := strings.Split(attrRole[1], ",")
				node.classes = append(node.classes, classes...)
			}
		}
	}
	return node
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

//
// indexByteUnescape find the index of the first unescaped byte `c` on
// slice of byte `in`.
// It will return nil and -1 if no unescape byte `c` found.
//
func indexByteUnescape(in []byte, c byte) (out []byte, idx int) {
	var (
		isEsc bool
	)
	out = make([]byte, 0, len(in))
	for x := 0; x < len(in); x++ {
		if in[x] == '\\' {
			if isEsc {
				out = append(out, '\\')
				isEsc = false
			} else {
				isEsc = true
			}
			continue
		}
		if in[x] == c {
			if isEsc {
				out = append(out, in[x])
				isEsc = false
				continue
			}
			return out, x
		}
		out = append(out, in[x])
	}
	return nil, -1
}

func indexUnescape(in []byte, token []byte) (out []byte, idx int) {
	tokenLen := len(token)
	if tokenLen > len(in) {
		return nil, -1
	}

	var (
		isEsc bool
	)
	out = make([]byte, 0, len(in))
	for x := 0; x < len(in); x++ {
		if in[x] == '\\' {
			if isEsc {
				out = append(out, '\\')
				isEsc = false
			} else {
				isEsc = true
			}
			continue
		}
		if in[x] == token[0] {
			if isEsc {
				out = append(out, in[x])
				isEsc = false
				continue
			}
			tmp := in[x:]
			if len(tmp) < tokenLen {
				return nil, -1
			}
			if bytes.Equal(tmp[:tokenLen], token) {
				return out, x
			}
		}
		out = append(out, in[x])
	}
	return nil, -1
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

func isEndFormat(prev, next byte) bool {
	if ascii.IsSpace(prev) || ascii.IsAlnum(next) {
		return false
	}
	return true
}
