// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"

	"github.com/shuLhan/share/lib/ascii"
)

//
// inlineParser is the one that responsible to parse text that contains inline
// markup (bold, italic, etc.) into tree.
//
type inlineParser struct {
	container *element
	current   *element
	content   []byte
	doc       *Document
	x         int
	state     *inlineParserState
	prev      byte
	c         byte
	nextc     byte
	nextcc    byte
	isEscaped bool
}

func newInlineParser(doc *Document, content []byte) (pi *inlineParser) {
	pi = &inlineParser{
		container: &element{
			kind: elKindText,
		},
		content: content,
		doc:     doc,
		state:   &inlineParserState{},
	}
	pi.current = pi.container

	return pi
}

func (pi *inlineParser) do() {
	for pi.x < len(pi.content) {
		pi.c = pi.content[pi.x]
		if pi.x+1 == len(pi.content) {
			pi.nextc = 0
		} else {
			pi.nextc = pi.content[pi.x+1]
		}
		if pi.x+2 >= len(pi.content) {
			pi.nextcc = 0
		} else {
			pi.nextcc = pi.content[pi.x+2]
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
			if pi.prev == ' ' && pi.nextc == '\n' {
				pi.current.backTrimSpace()
				pi.current.WriteString("<br>\n")
				pi.x += 2
				pi.prev = 0
				continue
			}
			if pi.parsePassthrough() {
				continue
			}
		} else if pi.c == ':' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if !ascii.IsSpace(pi.nextc) {
				if pi.parseMacro() {
					continue
				}
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
					elKindSymbolQuoteDoubleBegin)
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
					elKindSymbolQuoteSingleBegin)
				if ok {
					continue
				}
			}
			if ascii.IsAlpha(pi.prev) {
				pi.current.WriteString(htmlSymbolApostrophe)
				pi.x++
				pi.prev = pi.c
				continue
			}
		} else if pi.c == '*' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '*' {
				if pi.parseFormatUnconstrained(
					[]byte("**"),
					elKindUnconstrainedBold,
					elKindTextBold,
					styleTextBold) {
					continue
				}
			}
			if pi.parseFormat(elKindTextBold, styleTextBold) {
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
					elKindUnconstrainedItalic,
					elKindTextItalic,
					styleTextItalic) {
					continue
				}
			}
			if pi.parseFormat(elKindTextItalic, styleTextItalic) {
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
					elKindUnconstrainedMono,
					elKindTextMono,
					styleTextMono) {
					continue
				}
			}
			if pi.nextc == '"' {
				if pi.parseQuoteEnd([]byte("`\""),
					elKindSymbolQuoteDoubleEnd) {
					continue
				}
			}
			if pi.nextc == '\'' {
				if pi.parseQuoteEnd([]byte("`'"),
					elKindSymbolQuoteSingleEnd) {
					continue
				}

				// This is an aposthrope
				pi.current.WriteString(symbolQuoteSingleEnd)
				pi.x += 2
				pi.prev = 0
				continue
			}
			if pi.parseFormat(elKindTextMono, styleTextMono) {
				continue
			}
		} else if pi.c == '[' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '[' {
				if pi.parseInlineID() {
					continue
				}
			} else if pi.nextc == '#' {
				if pi.parseInlineIDShort() {
					continue
				}
			}
		} else if pi.c == '#' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			// Do we have the beginning?
			if pi.state.has(elKindInlineIDShort) {
				pi.terminate(elKindInlineIDShort, 0)
				pi.x++
				pi.prev = 0
				continue
			}
		} else if pi.c == '<' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '<' {
				if pi.parseCrossRef() {
					continue
				}
			} else if pi.nextc == '-' {
				pi.current.WriteString(htmlSymbolSingleLeftArrow)
				pi.x += 2
				pi.prev = pi.nextc
				continue
			} else if pi.nextc == '=' {
				pi.current.WriteString(htmlSymbolDoubleLeftArrow)
				pi.x += 2
				pi.prev = pi.nextc
				continue
			}
			pi.current.WriteString(htmlSymbolLessthan)
			pi.x++
			pi.prev = pi.c
			continue
		} else if pi.c == '>' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			pi.current.WriteString(htmlSymbolGreaterthan)
			pi.x++
			pi.prev = pi.c
			continue
		} else if pi.c == '&' {
			if ascii.IsSpace(pi.prev) && ascii.IsSpace(pi.nextc) {
				pi.current.WriteString(htmlSymbolAmpersand)
				pi.x += 2
				pi.prev = pi.nextc
				continue
			}
		} else if pi.c == '{' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			content, ok := parseAttrRef(pi.doc, pi.content, pi.x)
			if ok {
				pi.content = content
				pi.x = 0
				pi.prev = 0
				continue
			}
		} else if pi.c == '-' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.prev != '-' {
				if pi.nextc == '-' && pi.nextcc != '-' {
					pi.current.WriteString(htmlSymbolEmdash)
					pi.x += 2
					pi.prev = pi.nextc
					continue
				}
			} else if pi.nextc == '>' {
				pi.current.WriteString(htmlSymbolSingleRightArrow)
				pi.x += 2
				pi.prev = pi.nextc
				continue
			}
		} else if pi.c == '=' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '>' {
				pi.current.WriteString(htmlSymbolDoubleRightArrow)
				pi.x += 2
				pi.prev = pi.nextc
				continue
			}
		} else if pi.c == '.' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			if pi.nextc == '.' && pi.nextcc == '.' {
				pi.current.WriteString(htmlSymbolEllipsis)
				pi.x += 3
				pi.prev = pi.c
				continue
			}
		} else if pi.c == '(' {
			if pi.isEscaped {
				pi.escape()
				continue
			}
			var isReplaced bool
			raw, _ := indexByteUnescape(pi.content[pi.x+1:], ')')
			if len(raw) == 1 {
				if raw[0] == 'C' {
					pi.current.WriteString(htmlSymbolCopyright)
					isReplaced = true
				} else if raw[0] == 'R' {
					pi.current.WriteString(htmlSymbolRegistered)
					isReplaced = true
				}
			} else if len(raw) == 2 {
				if bytes.Equal(raw, []byte("TM")) {
					pi.current.WriteString(htmlSymbolTrademark)
					isReplaced = true
				}
			}
			if isReplaced {
				pi.x += len(raw) + 2
				pi.prev = ')'
				continue
			}
		}
		pi.current.WriteByte(pi.c)
		pi.x++
		pi.prev = pi.c
	}

	// Remove any trailing spaces only if the el is not passthrough.
	if !(pi.current.kind == elKindPassthrough ||
		pi.current.kind == elKindPassthroughDouble ||
		pi.current.kind == elKindPassthroughTriple) {
		pi.current.backTrimSpace()
	}
	pi.container.removeLastIfEmpty()
}

func (pi *inlineParser) escape() {
	pi.isEscaped = false
	pi.current.WriteByte(pi.c)
	pi.x++
	pi.prev = pi.c
}

func (pi *inlineParser) getBackMacroName() (macroName string, lastc byte) {
	raw := pi.current.raw
	start := len(raw) - 1
	for start >= 0 {
		if !ascii.IsAlpha(raw[start]) {
			return string(raw[start+1:]), raw[start]
		}
		start--
	}
	return string(raw), 0
}

func (pi *inlineParser) parseCrossRef() bool {
	raw := pi.content[pi.x+2:]
	raw, idx := indexUnescape(raw, []byte(">>"))
	if idx < 0 {
		return false
	}

	var (
		href  string
		label string
		title string
		ok    bool
	)

	parts := bytes.Split(raw, []byte(","))
	if len(parts) >= 2 {
		label = string(bytes.TrimSpace(parts[1]))
	}

	if isRefTitle(parts[0]) {
		// Get ID by title.
		href, ok = pi.doc.titleID[string(parts[0])]
		if ok {
			if len(label) == 0 {
				label = string(parts[0])
			}
		} else {
			// Store the label for cross reference later.
			title = string(parts[0])
		}
	} else if isValidID(parts[0]) {
		href = string(parts[0])
		if len(label) == 0 {
			anchor := pi.doc.anchors[href]
			if anchor != nil {
				label = anchor.label
			}
		}
	} else {
		return false
	}

	// The ID field will we non-empty if href is empty, it will be
	// revalidated later when rendered.
	elCrossRef := &element{
		elementAttribute: elementAttribute{
			Attrs: map[string]string{
				attrNameHref:  href,
				attrNameTitle: title,
			},
		},
		kind: elKindCrossReference,
		raw:  []byte(label),
	}
	pi.current.addChild(elCrossRef)
	el := &element{
		kind: elKindText,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2 + len(raw) + 2
	pi.prev = 0
	return true
}

//
// parseInlineID parse the ID and optional label between "[[" "]]".
//
func (pi *inlineParser) parseInlineID() bool {
	// Check if we have termination.
	raw := pi.content[pi.x+2:]
	raw, idx := indexUnescape(raw, []byte("]]"))
	if idx < 0 {
		return false
	}
	id, label := parseIDLabel(raw)
	if len(id) == 0 {
		return false
	}

	stringID := pi.doc.registerAnchor(string(id), string(label))

	el := &element{
		elementAttribute: elementAttribute{
			ID: stringID,
		},
		kind: elKindInlineID,
	}
	pi.current.backTrimSpace()
	pi.current.addChild(el)
	el = &element{
		kind: elKindText,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2 + len(raw) + 2
	pi.prev = 0
	return true
}

//
// parseInlineIDShort parse the ID and optional label between "[#", "]#", and
// "#".
//
func (pi *inlineParser) parseInlineIDShort() bool {
	// Check if we have term at the end.
	raw := pi.content[pi.x+2:]
	id, idx := indexUnescape(raw, []byte("]#"))
	if idx < 0 {
		return false
	}

	if !isValidID(id) {
		return false
	}

	// Check if we have "#"
	_, idx = indexByteUnescape(raw[idx+2:], '#')
	if idx < 0 {
		return false
	}

	stringID := pi.doc.registerAnchor(string(id), "")

	el := &element{
		elementAttribute: elementAttribute{
			ID: stringID,
		},
		kind: elKindInlineIDShort,
	}
	pi.state.push(elKindInlineIDShort)
	pi.current.backTrimSpace()
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2 + len(id) + 2
	return true
}

//
// parseQuoteBegin check if the double quote curve ("`) is valid (does not
// followed by space) and has an end (`").
//
func (pi *inlineParser) parseQuoteBegin(quoteEnd []byte, kind int) bool {
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
	el := &element{
		kind: kind,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2
	pi.prev = 0
	return true
}

func (pi *inlineParser) parseQuoteEnd(quoteEnd []byte, kind int) bool {
	if ascii.IsSpace(pi.prev) {
		// This is not the end that we looking for.
		return false
	}
	el := &element{
		kind: kind,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2
	pi.prev = 0
	return true
}

func (pi *inlineParser) parseFormat(kind int, style int64) bool {
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

	el := &element{
		kind: kind,
	}
	pi.current.addChild(el)
	pi.state.push(kind)
	pi.current = el
	pi.prev = 0
	pi.x++
	return true
}

func (pi *inlineParser) parseFormatUnconstrained(
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
		el := &element{
			kind: kindUnconstrained,
		}
		pi.current.addChild(el)
		pi.state.push(kindUnconstrained)
		pi.current = el
		pi.prev = 0
		pi.x += 2
		return true
	}

	return false
}

func (pi *inlineParser) parseInlineImage() *element {
	content := pi.content[pi.x+1:]
	_, idx := indexByteUnescape(content, ']')
	if idx < 0 {
		return nil
	}

	lineImage := content[:idx+1]
	elImage := &element{
		elementAttribute: elementAttribute{
			Attrs: make(map[string]string),
		},
		kind: elKindInlineImage,
	}
	if elImage.parseBlockImage(pi.doc, lineImage) {
		pi.x += idx + 2
		pi.prev = 0
		return elImage
	}
	return nil
}

func (pi *inlineParser) parseMacro() bool {
	name, lastc := pi.getBackMacroName()
	if lastc == '\\' || len(name) == 0 {
		return false
	}

	switch name {
	case "":
		return false
	case macroFTP, macroHTTPS, macroHTTP, macroIRC, macroLink, macroMailto:
		el := pi.parseURL(name)
		if el == nil {
			return false
		}

		pi.current.raw = pi.current.raw[:len(pi.current.raw)-len(name)]

		pi.current.addChild(el)
		el = &element{
			kind: elKindText,
		}
		pi.current.addChild(el)
		pi.current = el
		return true
	case macroImage:
		el := pi.parseInlineImage()
		if el == nil {
			return false
		}

		pi.current.raw = pi.current.raw[:len(pi.current.raw)-len(name)]

		pi.current.addChild(el)
		el = &element{
			kind: elKindText,
		}
		pi.current.addChild(el)
		pi.current = el
		return true
	}
	return false
}

func (pi *inlineParser) parsePassthrough() bool {
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

	el := &element{
		kind: elKindPassthrough,
		raw:  pass,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += x + 2
	pi.prev = 0
	return true
}

func (pi *inlineParser) parsePassthroughDouble() bool {
	raw := pi.content[pi.x+2:]

	// Check if we have "++" at the end.
	raw, idx := indexUnescape(raw, []byte("++"))
	if idx >= 0 {
		el := &element{
			kind: elKindPassthroughDouble,
			raw:  raw,
		}
		pi.current.addChild(el)
		pi.current = el
		pi.x += idx + 4
		pi.prev = 0
		return true
	}

	return false
}

func (pi *inlineParser) parsePassthroughTriple() bool {
	raw := pi.content[pi.x+3:]

	// Check if we have "+++" at the end.
	raw, idx := indexUnescape(raw, []byte("+++"))
	if idx >= 0 {
		el := &element{
			kind: elKindPassthroughTriple,
			raw:  raw,
		}
		pi.current.addChild(el)
		pi.current = el
		pi.x += idx + 6
		pi.prev = 0
		return true
	}
	return false
}

func (pi *inlineParser) parseSubscript() bool {
	var (
		prev byte
		raw  = pi.content[pi.x+1:]
	)
	for x := 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			el := &element{
				kind: elKindTextSubscript,
				raw:  raw[:x],
			}
			pi.current.addChild(el)

			el = &element{
				kind: elKindText,
			}
			pi.current.addChild(el)
			pi.current = el

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

func (pi *inlineParser) parseSuperscript() bool {
	var (
		prev byte
		raw  = pi.content[pi.x+1:]
	)
	for x := 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			el := &element{
				kind: elKindTextSuperscript,
				raw:  raw[:x],
			}
			pi.current.addChild(el)

			el = &element{
				kind: elKindText,
			}
			pi.current.addChild(el)
			pi.current = el

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
func (pi *inlineParser) parseURL(scheme string) (el *element) {
	var (
		x   int
		c   byte
		uri []byte
	)
	if scheme != macroLink {
		uri = []byte(scheme)
		uri = append(uri, ':')
	}

	el = &element{
		elementAttribute: elementAttribute{
			Attrs: make(map[string]string),
		},
		kind: elKindURL,
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
			el.addRole(attrValueBare)
		}
		if c == '.' || c == ',' || c == ';' {
			uri = uri[:len(uri)-1]
			pi.prev = 0
			pi.x += x
		} else {
			pi.x += x + 1
			pi.prev = c
		}
	}

	uri = applySubstitutions(pi.doc, uri)
	el.Attrs[attrNameHref] = string(uri)

	if c != '[' {
		el.raw = uri
		return el
	}

	_, idx := indexByteUnescape(content[x:], ']')
	if idx < 0 {
		return nil
	}

	pi.x += x + idx + 2
	pi.prev = 0

	attr := content[x : x+idx+1]
	el.style = styleLink
	el.parseElementAttribute(attr)
	if len(el.Attrs) == 0 {
		// empty "[]"
		el.raw = uri
		return el
	}
	if len(el.rawStyle) >= 1 {
		l := len(el.rawStyle)
		if el.rawStyle[l-1] == '^' {
			el.Attrs[attrNameTarget] = attrValueBlank
			el.rawStyle = el.rawStyle[:l-1]
			el.Attrs[attrNameRel] = attrValueNoopener
		}
		child := parseInlineMarkup(pi.doc, []byte(el.rawStyle))
		el.addChild(child)
	}
	return el
}

func (pi *inlineParser) terminate(kind int, style int64) {
	el := pi.current
	stateTmp := &inlineParserState{}
	for el.parent != nil {
		if el.kind == kind {
			pi.state.pop()
			el.style |= style
			break
		}
		if el.kind == elKindTextBold && el.style == 0 {
			el.style = styleTextBold
			stateTmp.push(pi.state.pop())
		}
		if el.kind == elKindTextItalic && el.style == 0 {
			el.style = styleTextItalic
			stateTmp.push(pi.state.pop())
		}
		if el.kind == elKindTextMono && el.style == 0 {
			el.style = styleTextMono
			stateTmp.push(pi.state.pop())
		}
		el = el.parent
	}
	if el.parent != nil {
		el = el.parent
	}
	for k := stateTmp.pop(); k != 0; k = stateTmp.pop() {
		child := &element{
			kind: k,
		}
		el.addChild(child)
		el = child
		pi.state.push(k)
	}
	child := &element{
		kind: elKindText,
	}
	el.addChild(child)
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
