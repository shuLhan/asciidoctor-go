// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"

	"github.com/shuLhan/share/lib/ascii"
)

// inlineParser is the one that responsible to parse text that contains inline
// markup (bold, italic, etc.) into tree.
type inlineParser struct {
	container *element
	current   *element

	doc   *Document
	state *inlineParserState

	content []byte

	x      int
	prev   byte
	c      byte
	nextc  byte
	nextcc byte

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
	var (
		vbytes []byte
		ok     bool
	)

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
				ok = pi.parseQuoteBegin([]byte("`\""), elKindSymbolQuoteDoubleBegin)
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
				ok = pi.parseQuoteBegin([]byte("`'"), elKindSymbolQuoteSingleBegin)
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
					[]byte(`**`),
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
					[]byte(`__`),
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

			vbytes, ok = parseAttrRef(pi.doc, pi.content, pi.x)
			if ok {
				pi.content = vbytes
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
					if ascii.IsSpace(pi.prev) && ascii.IsSpace(pi.nextcc) {
						pi.current.backTrimSpace()
						pi.current.WriteString(htmlSymbolThinSpace)
					}
					pi.current.WriteString(htmlSymbolEmdash)
					if ascii.IsSpace(pi.nextcc) {
						pi.current.WriteString(htmlSymbolThinSpace)
						pi.x++
					}
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
				pi.current.WriteString(htmlSymbolZeroWidthSpace)
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
			vbytes, _ = indexByteUnescape(pi.content[pi.x+1:], ')')
			if len(vbytes) == 1 {
				if vbytes[0] == 'C' {
					pi.current.WriteString(htmlSymbolCopyright)
					isReplaced = true
				} else if vbytes[0] == 'R' {
					pi.current.WriteString(htmlSymbolRegistered)
					isReplaced = true
				}
			} else if len(vbytes) == 2 {
				if bytes.Equal(vbytes, []byte(`TM`)) {
					pi.current.WriteString(htmlSymbolTrademark)
					isReplaced = true
				}
			}
			if isReplaced {
				pi.x += len(vbytes) + 2
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

func (pi *inlineParser) parseCrossRef() bool {
	var (
		raw []byte = pi.content[pi.x+2:]
		idx int
	)

	raw, idx = indexUnescape(raw, []byte(`>>`))
	if idx < 0 {
		return false
	}

	var (
		elCrossRef *element
		el         *element
		href       string
		label      string
		parts      [][]byte
	)

	parts = bytes.Split(raw, []byte(`,`))
	href = string(parts[0])
	if len(parts) >= 2 {
		label = string(bytes.TrimSpace(parts[1]))
	}

	// Set attribute href to the first part, we will revalidated later
	// when rendering the element.
	elCrossRef = &element{
		elementAttribute: elementAttribute{
			Attrs: map[string]string{
				attrNameHref: href,
			},
		},
		kind: elKindCrossReference,
		raw:  []byte(label),
	}
	pi.current.addChild(elCrossRef)
	el = &element{
		kind: elKindText,
	}
	pi.current.addChild(el)
	pi.current = el
	pi.x += 2 + len(raw) + 2
	pi.prev = 0
	return true
}

// parseInlineID parse the ID and optional label between "[[" "]]".
func (pi *inlineParser) parseInlineID() bool {
	var (
		raw []byte = pi.content[pi.x+2:]

		el       *element
		stringID string
		id       []byte
		label    []byte
		idx      int
	)

	// Check if we have termination.
	raw, idx = indexUnescape(raw, []byte("]]"))
	if idx < 0 {
		return false
	}
	id, label = parseIDLabel(raw)
	if len(id) == 0 {
		return false
	}

	stringID = pi.doc.registerAnchor(string(id), string(label))

	el = &element{
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

// parseInlineIDShort parse the ID and optional label between "[#", "]#", and
// "#".
func (pi *inlineParser) parseInlineIDShort() bool {
	var (
		raw []byte = pi.content[pi.x+2:]

		el       *element
		stringID string
		id       []byte
		idx      int
	)

	// Check if we have term at the end.
	id, idx = indexUnescape(raw, []byte(`]#`))
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

	stringID = pi.doc.registerAnchor(string(id), ``)

	el = &element{
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

// parseQuoteBegin check if the double quote curve ("`) is valid (does not
// followed by space) and has an end (`").
func (pi *inlineParser) parseQuoteBegin(quoteEnd []byte, kind int) bool {
	if pi.x+2 >= len(pi.content) {
		return false
	}

	var c byte = pi.content[pi.x+2]
	if ascii.IsSpace(c) {
		return false
	}

	var (
		raw []byte = pi.content[pi.x+2:]
		idx int    = bytes.LastIndex(raw, quoteEnd)
	)

	if idx < 0 {
		return false
	}
	if ascii.IsSpace(raw[idx-1]) || raw[idx-1] == '\\' {
		return false
	}

	var el = &element{
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
	var el = &element{
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

	var (
		raw []byte = pi.content[pi.x+1:]

		el     *element
		idx    int
		prevc  byte
		nextc  byte
		hasEnd bool
	)

	_, idx = indexByteUnescape(raw, pi.c)
	for idx >= 0 {
		prevc = 0
		nextc = 0

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

	el = &element{
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
	var (
		raw []byte = pi.content[pi.x+2:]
		el  *element
	)

	if bytes.Contains(raw, terms) {
		el = &element{
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
	var (
		content []byte = pi.content[pi.x+1:]

		elImage   *element
		lineImage []byte
		idx       int
	)

	_, idx = indexByteUnescape(content, ']')
	if idx < 0 {
		return nil
	}

	lineImage = content[:idx+1]
	elImage = &element{
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
	var (
		el   *element
		name string
		n    int
	)

	name = pi.parseMacroName(pi.current.raw)
	if len(name) == 0 {
		return false
	}

	switch name {
	case macroFootnote:
		el, n = pi.parseMacroFootnote(pi.content[pi.x+1:])
		if el == nil {
			return false
		}

		pi.x += n
		pi.prev = 0

	case macroFTP, macroHTTPS, macroHTTP, macroIRC, macroLink, macroMailto:
		el = pi.parseURL(name)
		if el == nil {
			return false
		}

	case macroImage:
		el = pi.parseInlineImage()
		if el == nil {
			return false
		}
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

	var el = &element{
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
	var (
		raw []byte = pi.content[pi.x+2:]
		idx int
		el  *element
	)

	// Check if we have "++" at the end.
	raw, idx = indexUnescape(raw, []byte("++"))
	if idx >= 0 {
		el = &element{
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
	var (
		raw []byte = pi.content[pi.x+3:]
		idx int
		el  *element
	)

	// Check if we have "+++" at the end.
	raw, idx = indexUnescape(raw, []byte(`+++`))
	if idx >= 0 {
		el = &element{
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
		raw = pi.content[pi.x+1:]

		el   *element
		x    int
		prev byte
	)
	for x = 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			el = &element{
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
		raw = pi.content[pi.x+1:]

		el   *element
		x    int
		prev byte
	)
	for x = 0; x < len(raw); x++ {
		if raw[x] == pi.c {
			if prev == '\\' {
				continue
			}
			el = &element{
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

// parseURL parser the URL, an optional text, optional attribute for target,
// and optional role.
//
// The current state of p.x is equal to ":".
func (pi *inlineParser) parseURL(scheme string) (el *element) {
	var (
		x       int
		idx     int
		c       byte
		uri     []byte
		content []byte
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

	content = pi.content[pi.x+1:]
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

	_, idx = indexByteUnescape(content[x:], ']')
	if idx < 0 {
		return nil
	}

	pi.x += x + idx + 2
	pi.prev = 0

	var attr []byte = content[x : x+idx+1]
	el.style = styleLink
	el.parseElementAttribute(attr)
	if len(el.Attrs) == 0 {
		// empty "[]"
		el.raw = uri
		return el
	}
	if len(el.rawStyle) >= 1 {
		var (
			l int = len(el.rawStyle)

			child *element
		)

		if el.rawStyle[l-1] == '^' {
			el.Attrs[attrNameTarget] = attrValueBlank
			el.rawStyle = el.rawStyle[:l-1]
			el.Attrs[attrNameRel] = attrValueNoopener
		}
		child = parseInlineMarkup(pi.doc, []byte(el.rawStyle))
		el.addChild(child)
	}
	return el
}

func (pi *inlineParser) terminate(kind int, style int64) {
	var (
		el       *element = pi.current
		stateTmp          = &inlineParserState{}
	)

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

	var (
		child *element
		k     int
	)
	for k = stateTmp.pop(); k != 0; k = stateTmp.pop() {
		child = &element{
			kind: k,
		}
		el.addChild(child)
		el = child
		pi.state.push(k)
	}
	child = &element{
		kind: elKindText,
	}
	el.addChild(child)
	pi.current = child
}

// indexByteUnescape find the index of the first unescaped byte `c` on
// slice of byte `in`.
// It will return nil and -1 if no unescape byte `c` found.
func indexByteUnescape(in []byte, c byte) (out []byte, idx int) {
	var (
		x     int
		isEsc bool
	)
	out = make([]byte, 0, len(in))
	for x = 0; x < len(in); x++ {
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
	var (
		tokenLen int = len(token)

		tmp   []byte
		x     int
		isEsc bool
	)

	if tokenLen > len(in) {
		return nil, -1
	}

	out = make([]byte, 0, len(in))
	for x = 0; x < len(in); x++ {
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
			tmp = in[x:]
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
