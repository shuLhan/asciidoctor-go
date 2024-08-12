// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

const debugLevel = 0

type documentParser struct {
	doc      *Document
	lines    [][]byte
	lineNum  int
	prevKind int
	kind     int
}

func newDocumentParser(doc *Document, content []byte) (docp *documentParser) {
	docp = &documentParser{
		doc: doc,
	}

	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	docp.lines = bytes.Split(content, []byte("\n"))

	var (
		wspaces = "\t\n\v\f\r \x85\xA0"

		line []byte
		x    int
	)
	for x, line = range docp.lines {
		docp.lines[x] = bytes.TrimRight(line, wspaces)
	}

	return docp
}

func parseSub(parentDoc *Document, content []byte) (subdoc *Document) {
	var (
		docp *documentParser
		k    string
		v    string
	)

	subdoc = newDocument()

	for k, v = range parentDoc.Attributes {
		subdoc.Attributes[k] = v
	}

	docp = newDocumentParser(subdoc, content)

	docp.parseBlock(subdoc.content, 0)

	return subdoc
}

// consumeLinesUntil given an element el, consume all lines until we found
// a line with kind match with term or match with one in terms.
func (docp *documentParser) consumeLinesUntil(el *element, term int, terms []int) (line []byte) {
	var (
		logp = `consumeLinesUntil`

		elInclude    *elementInclude
		spaces       []byte
		t            int
		ok           bool
		allowComment bool
	)

	if term == elKindBlockListing || term == elKindBlockListingNamed ||
		term == elKindLiteralParagraph {
		allowComment = true
	}
	for {
		spaces, line, ok = docp.line(logp)
		if !ok {
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if docp.kind == lineKindComment {
			if allowComment {
				el.Write(line)
				el.WriteByte('\n')
			}
			continue
		}
		if docp.kind == term {
			el.raw = bytes.TrimRight(el.raw, " \n")
			return nil
		}
		for _, t = range terms {
			if t == docp.kind {
				return line
			}
		}
		if docp.kind == lineKindInclude {
			elInclude = parseInclude(docp.doc, line)
			if elInclude == nil {
				el.Write(line)
				el.WriteByte('\n')
				line = nil
				continue
			}
			// Include the content of file into the current
			// document.
			docp.include(elInclude)
			line = nil
			continue
		}
		if el.kind == elKindBlockPassthrough ||
			el.kind == elKindBlockListing ||
			el.kind == elKindBlockLiteral {
			if el.kind != elKindTable {
				el.Write(spaces)
			}
		} else if el.kind == elKindParagraph && len(spaces) > 0 {
			el.WriteByte(' ')
		}
		el.Write(line)
		el.WriteByte('\n')
	}
	return line
}

// hasPreamble will return true if the contents contains preamble, indicated
// by the first section that found after current line.
func (docp *documentParser) hasPreamble() bool {
	var (
		start = docp.lineNum

		notEmtpy int
		line     []byte
		kind     int
	)
	for ; start < len(docp.lines); start++ {
		line = docp.lines[start]
		if len(line) == 0 {
			continue
		}
		kind, _, _ = whatKindOfLine(line)
		if kind == elKindSectionL1 || kind == elKindSectionL2 ||
			kind == elKindSectionL3 || kind == elKindSectionL4 ||
			kind == elKindSectionL5 ||
			kind == lineKindID ||
			kind == lineKindIDShort {
			return notEmtpy > 0
		}
		notEmtpy++
	}
	return false
}

func (docp *documentParser) include(el *elementInclude) {
	var content = bytes.ReplaceAll(el.content, []byte("\r\n"), []byte("\n"))

	content = bytes.TrimRight(content, "\n")

	var (
		includedLines = bytes.Split(content, []byte("\n"))
		newLines      = make([][]byte, 0, len(docp.lines)+len(includedLines))

		line []byte
	)

	// Do not add the "include" directive
	docp.lineNum--
	newLines = append(newLines, docp.lines[:docp.lineNum]...)
	newLines = append(newLines, includedLines...)
	newLines = append(newLines, docp.lines[docp.lineNum+1:]...)
	docp.lines = newLines

	if debugLevel >= 2 {
		for _, line = range includedLines {
			fmt.Printf("%s\n", line)
		}
	}
}

// line return the next line in the content of raw document.
// It will return ok as false if there are no more line.
func (docp *documentParser) line(logp string) (spaces, line []byte, ok bool) {
	docp.prevKind = docp.kind

	if docp.lineNum >= len(docp.lines) {
		return nil, nil, false
	}

	line = docp.lines[docp.lineNum]
	if debugLevel >= 2 {
		fmt.Printf("line %3d: %s: %s\n", docp.lineNum, logp, line)
	}
	docp.lineNum++

	docp.kind, spaces, line = whatKindOfLine(line)
	return spaces, line, true
}

// parseAttribute parse document attribute and return its key and optional
// value.
func (docp *documentParser) parseAttribute(line []byte, strict bool) (key, value string, ok bool) {
	var (
		bb bytes.Buffer
		p  int
		x  int
	)

	if !(ascii.IsAlnum(line[1]) || line[1] == '_') {
		return ``, ``, false
	}

	bb.WriteByte(line[1])
	x = 2
	for ; x < len(line); x++ {
		if line[x] == ':' {
			break
		}
		if ascii.IsAlnum(line[x]) || line[x] == '_' ||
			line[x] == '-' || line[x] == '!' {
			bb.WriteByte(line[x])
			continue
		}
		if strict {
			return ``, ``, false
		}
	}
	if x == len(line) {
		return ``, ``, false
	}

	key = bb.String()

	line = line[x+1:]
	p = len(line)
	if p > 0 && line[p-1] == '\\' {
		bb.Reset()
		line = line[:p-1]
		bb.Write(line)
		docp.parseMultiline(&bb)
		line = bb.Bytes()
	}

	line = bytes.TrimSpace(line)
	value = string(line)

	return key, value, true
}

// parseMultiline multiline value where each line end with `\`.
func (docp *documentParser) parseMultiline(out io.Writer) {
	var (
		isMultiline = true

		line []byte
		p    int
	)
	for isMultiline && docp.lineNum < len(docp.lines) {
		line = docp.lines[docp.lineNum]
		p = len(line)

		if line[p-1] == '\\' {
			line = line[:p-1]
		} else {
			isMultiline = false
		}
		_, _ = out.Write(line)
		docp.lineNum++
	}
}

func (docp *documentParser) parseBlock(parent *element, term int) {
	var (
		logp = `parseBlock`
		el   = &element{}

		line   []byte
		isTerm bool
		ok     bool
	)
	for !isTerm {
		if len(line) == 0 {
			_, line, ok = docp.line(logp)
			if !ok {
				return
			}
		}

		switch docp.kind {
		case term:
			isTerm = true
			continue
		case lineKindEmpty:
			line = nil
			continue
		case lineKindBlockComment:
			docp.parseIgnoreCommentBlock()
			line = nil
			continue
		case lineKindComment:
			line = nil
			continue
		case lineKindHorizontalRule:
			el.kind = docp.kind
			parent.addChild(el)
			el = &element{}
			line = nil
			continue

		case lineKindID:
			if parent.kind == elKindPreamble {
				docp.kind = lineKindEmpty
				docp.prevKind = lineKindEmpty
				docp.lineNum--
				isTerm = true
				continue
			}
			var (
				idLabel   = line[2 : len(line)-2]
				id, label = parseIDLabel(idLabel)
			)
			if len(id) > 0 {
				el.ID = docp.doc.registerAnchor(string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindIDShort:
			if parent.kind == elKindPreamble {
				docp.kind = lineKindEmpty
				docp.prevKind = lineKindEmpty
				docp.lineNum--
				isTerm = true
				continue
			}

			var (
				id    = line[2 : len(line)-1]
				label []byte
			)

			id, label = parseIDLabel(id)

			if len(id) > 0 {
				el.ID = docp.doc.registerAnchor(string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindInclude:
			var elInclude = parseInclude(docp.doc, []byte(line))

			if elInclude == nil {
				el.Write(line)
				el.WriteByte('\n')
				line = nil
				continue
			}
			// Include the content of file into the current
			// document.
			docp.include(elInclude)
			line = nil
			continue

		case lineKindPageBreak:
			el.kind = docp.kind
			parent.addChild(el)
			el = &element{}
			line = nil
			continue

		case lineKindAttribute:
			var (
				key, value, ok = docp.parseAttribute(line, false)
			)
			if ok {
				if key == attrNameIcons {
					if el.Attrs == nil {
						el.Attrs = make(map[string]string)
					}
					el.Attrs[key] = value
				} else {
					docp.doc.Attributes.apply(key, value)
					parent.addChild(&element{
						kind:  docp.kind,
						key:   key,
						value: value,
					})
				}
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindAttributeElement:
			el.parseElementAttribute(line)
			if el.style > 0 {
				if isStyleAdmonition(el.style) {
					el.setStyleAdmonition(el.rawStyle)
				}
			}
			line = nil
			continue

		case lineKindStyleClass:
			el.parseStyleClass(line)
			line = nil
			continue

		case lineKindText, lineKindListContinue:
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindBlockTitle:
			el.rawTitle = string(line[1:])
			line = nil
			continue

		case lineKindAdmonition:
			el.kind = elKindParagraph
			el.style |= styleAdmonition
			el.parseLineAdmonition(line)
			line = docp.consumeLinesUntil(
				el,
				lineKindEmpty,
				[]int{
					term,
					elKindBlockListing,
					elKindBlockListingNamed,
					elKindBlockLiteral,
					elKindBlockLiteralNamed,
					lineKindListContinue,
				})
			el.parseInlineMarkup(docp.doc, elKindText)
			parent.addChild(el)
			el = new(element)
			continue

		case elKindSectionL1, elKindSectionL2, elKindSectionL3, elKindSectionL4, elKindSectionL5:
			if parent.kind == elKindPreamble {
				docp.kind = lineKindEmpty
				docp.prevKind = lineKindEmpty
				docp.lineNum--
				isTerm = true
				continue
			}

			if term == elKindBlockOpen {
				line = docp.parseParagraph(parent, el, line, term)
				parent.addChild(el)
				el = new(element)
				continue
			}

			el.kind = docp.kind
			// BUG: "= =a" could become "a", it should be "=a"
			el.Write(bytes.TrimLeft(line, "= \t"))

			var isDiscrete = el.style&styleSectionDiscrete > 0
			if isDiscrete {
				el.kind = elKindSectionDiscrete
				el.level = docp.kind
				el.parseSection(docp.doc, isDiscrete)
				parent.addChild(el)
				el = new(element)
				line = nil
				continue
			}

			var expParent = docp.kind - 1
			for parent.kind != expParent {
				parent = parent.parent
				if parent == nil {
					parent = docp.doc.content
					break
				}
			}
			el.parseSection(docp.doc, false)
			parent.addChild(el)
			parent = el
			el = new(element)
			line = nil
			continue

		case elKindLiteralParagraph:
			if el.isStyleAdmonition() {
				line = docp.parseParagraph(parent, el,
					line, term)
			} else {
				el.kind = docp.kind
				el.addRole(classNameLiteralBlock)
				el.Write(line)
				el.WriteByte('\n')
				line = docp.consumeLinesUntil(
					el,
					lineKindEmpty,
					[]int{
						term,
						elKindBlockListing,
						elKindBlockListingNamed,
						elKindBlockLiteral,
						elKindBlockLiteralNamed,
					})
				el.raw = applySubstitutions(docp.doc, el.raw)
			}
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockLiteral:
			el.kind = docp.kind
			el.addRole(classNameLiteralBlock)
			line = docp.consumeLinesUntil(el, docp.kind, nil)
			el.raw = applySubstitutions(docp.doc, el.raw)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockLiteralNamed:
			el.kind = docp.kind
			el.addRole(classNameLiteralBlock)
			line = docp.consumeLinesUntil(el, lineKindEmpty, nil)
			el.raw = applySubstitutions(docp.doc, el.raw)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockListing:
			el.kind = docp.kind
			el.addRole(classNameListingBlock)
			line = docp.consumeLinesUntil(el, docp.kind, nil)
			el.raw = applySubstitutions(docp.doc, el.raw)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockListingNamed:
			el.kind = docp.kind
			el.addRole(classNameListingBlock)
			line = docp.consumeLinesUntil(
				el,
				lineKindEmpty,
				[]int{
					elKindBlockListing,
					elKindBlockListingNamed,
					elKindBlockLiteral,
					elKindBlockLiteralNamed,
					lineKindListContinue,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockPassthrough:
			el.kind = docp.kind
			line = docp.consumeLinesUntil(el, docp.kind, nil)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindListOrderedItem:
			line = docp.parseListOrdered(parent, el.rawTitle, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindListUnorderedItem:
			line = docp.parseListUnordered(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindListDescriptionItem:
			line = docp.parseListDescription(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockImage:
			var lineImage = bytes.TrimRight(line[7:], " \t")
			if el.parseBlockImage(docp.doc, lineImage) {
				el.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, el, line, term)
			}
			parent.addChild(el)
			el = &element{}
			continue

		case elKindBlockOpen, elKindBlockExample, elKindBlockSidebar:
			el.kind = docp.kind
			docp.parseBlock(el, docp.kind)
			parent.addChild(el)
			el = new(element)
			line = nil
			continue

		case elKindBlockExcerpts:
			el.kind = docp.kind
			if el.isStyleVerse() {
				line = docp.consumeLinesUntil(
					el,
					docp.kind,
					[]int{
						term,
						elKindBlockListing,
						elKindBlockListingNamed,
						elKindBlockLiteral,
						elKindBlockLiteralNamed,
						lineKindListContinue,
					})
			} else {
				docp.parseBlock(el, docp.kind)
				line = nil
			}
			parent.addChild(el)
			el = new(element)
			continue

		case elKindBlockVideo:
			if el.parseBlockVideo(docp.doc, line) {
				el.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, el, line, term)
			}
			parent.addChild(el)
			el = new(element)
			continue

		case elKindBlockAudio:
			if el.parseBlockAudio(docp.doc, line) {
				el.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, el, line, term)
			}
			parent.addChild(el)
			el = new(element)
			continue

		case elKindMacroTOC:
			el.kind = docp.kind
			parent.addChild(el)
			el = new(element)

		case elKindTable:
			el.kind = docp.kind
			line = docp.consumeLinesUntil(el, docp.kind, nil)
			parent.addChild(el)
			el.postConsumeTable()
			el = &element{}
			continue
		}
		line = nil
	}
}

// parseHeader document header consist of title and optional authors,
// revision, and zero or more attributes.
//
// The document attributes can be in any order, but the author and revision
// MUST be in order.
//
//	DOC_HEADER  = [ "=" SP *ADOC_WORD LF
//	              [ DOC_AUTHORS LF
//	              [ DOC_REVISION LF ]]]
//	              (*DOC_ATTRIBUTE)
//	              LF
func (docp *documentParser) parseHeader() {
	var (
		logp = `parseHeader`
		line []byte
		ok   bool
	)

	line, ok = docp.skipCommentAndEmptyLine()
	if !ok {
		return
	}
	if docp.kind == elKindSectionL0 {
		docp.doc.header.Write(bytes.TrimSpace(line[2:]))
		docp.doc.Title.raw = string(docp.doc.header.raw)

		_, line, ok = docp.line(logp)
		if !ok {
			return
		}
		if docp.kind == lineKindText {
			docp.doc.rawAuthors = string(line)

			_, line, ok = docp.line(logp)
			if !ok {
				return
			}
			if docp.kind == lineKindText {
				docp.doc.rawRevision = string(line)
				line = nil
			}
		}
	}

	// Parse the rest of attributes until we found an empty line or
	// line with non-attribute.
	for {
		if line == nil {
			_, line, ok = docp.line(logp)
			if !ok {
				return
			}
		}
		if docp.kind == lineKindEmpty {
			return
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = nil
			continue
		}
		if docp.kind == lineKindComment {
			line = nil
			continue
		}
		if docp.kind == lineKindAttribute {
			var key, value string
			key, value, ok = docp.parseAttribute(line, false)
			if ok {
				docp.doc.Attributes.apply(key, value)
			}
			line = nil
			continue
		}
		docp.lineNum--
		break
	}
}

func (docp *documentParser) parseIgnoreCommentBlock() {
	var (
		logp = `parseIgnoreCommentBlock`

		line []byte
		ok   bool
	)

	for {
		_, line, ok = docp.line(logp)
		if !ok {
			return
		}
		if bytes.HasPrefix(line, []byte(`////`)) {
			return
		}
	}
}

// parseListBlock parse block after list continuation `+` until we found
// empty line or non-list line.
func (docp *documentParser) parseListBlock() (el *element, line []byte) {
	var (
		logp = `parseListBlock`

		ok bool
	)

	for {
		_, line, ok = docp.line(logp)
		if !ok {
			break
		}

		if docp.kind == lineKindAdmonition {
			el = &element{
				elementAttribute: elementAttribute{
					style: styleAdmonition,
				},
				kind: elKindParagraph,
			}
			el.parseLineAdmonition(line)
			line = docp.consumeLinesUntil(
				el,
				lineKindEmpty,
				[]int{
					elKindBlockListing,
					elKindBlockListingNamed,
					elKindBlockLiteral,
					elKindBlockLiteralNamed,
					lineKindListContinue,
				})
			el.parseInlineMarkup(docp.doc, elKindText)
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if docp.kind == lineKindComment {
			continue
		}
		if docp.kind == lineKindEmpty {
			return el, line
		}
		if docp.kind == lineKindListContinue {
			continue
		}
		if docp.kind == elKindLiteralParagraph {
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			el.Write(bytes.TrimLeft(line, " \t"))
			el.WriteByte('\n')
			line = docp.consumeLinesUntil(
				el,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			break
		}
		if docp.kind == lineKindText {
			el = &element{
				kind: elKindParagraph,
			}
			el.Write(line)
			el.WriteByte('\n')
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					elKindListDescriptionItem,
				})
			el.postParseParagraph(nil)
			el.parseInlineMarkup(docp.doc, elKindText)
			break
		}
		if docp.kind == elKindBlockListing {
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			docp.consumeLinesUntil(el, docp.kind, nil)
			el.raw = applySubstitutions(docp.doc, el.raw)
			line = nil
			break
		}
		if docp.kind == elKindBlockOpen {
			el = &element{
				kind: docp.kind,
			}
			docp.parseBlock(el, docp.kind)
			line = nil
			break
		}
		if docp.kind == elKindListOrderedItem {
			break
		}
		if docp.kind == elKindListUnorderedItem {
			break
		}
		if docp.kind == elKindListDescriptionItem {
			break
		}
	}
	return el, line
}

// parseListDescription parse the list description item, line that end with
// "::" WSP, and its content.
func (docp *documentParser) parseListDescription(parent, el *element, line []byte, term int) (got []byte) {
	var (
		logp = `parseListDescription`
		list = &element{
			elementAttribute: elementAttribute{
				style: el.style,
			},
			kind:     elKindListDescription,
			rawTitle: el.rawTitle,
		}
		listItem = &element{
			elementAttribute: elementAttribute{
				style: list.style,
			},
			kind: elKindListDescriptionItem,
		}

		ok bool
	)

	listItem.parseListDescriptionItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	line = nil
	for {
		if len(line) == 0 {
			_, line, ok = docp.line(logp)
			if !ok {
				break
			}
		}
		if docp.kind == term {
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = nil
			continue
		}
		if docp.kind == lineKindComment {
			line = nil
			continue
		}
		if docp.kind == lineKindInclude {
			var elInclude = parseInclude(docp.doc, line)
			if elInclude == nil {
				el.Write(line)
				el.WriteByte('\n')
				line = nil
				continue
			}
			// Include the content of file into the current
			// document.
			docp.include(elInclude)
			line = nil
			continue
		}
		if docp.kind == lineKindListContinue {
			el, line = docp.parseListBlock()
			if el != nil {
				listItem.addChild(el)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == elKindListOrderedItem {
			line = docp.parseListOrdered(listItem, ``, line, term)
			continue
		}
		if docp.kind == elKindListUnorderedItem {
			line = docp.parseListUnordered(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindListDescriptionItem {
			var el = &element{
				elementAttribute: elementAttribute{
					style: list.style,
				},
				kind: elKindListDescriptionItem,
			}
			el.parseListDescriptionItem(line)
			if listItem.level == el.level {
				list.addChild(el)
				listItem = el
				line = nil
				continue
			}

			var parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc,
						elKindListDescriptionItem)
					return line
				}
				parentListItem = parentListItem.parent
			}
			line = docp.parseListDescription(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			var el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			var el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockListing ||
			docp.kind == elKindBlockExample ||
			docp.kind == elKindBlockSidebar {
			break
		}
		if docp.kind == elKindSectionL1 ||
			docp.kind == elKindSectionL2 ||
			docp.kind == elKindSectionL3 ||
			docp.kind == elKindSectionL4 ||
			docp.kind == elKindSectionL5 ||
			docp.kind == lineKindAttributeElement ||
			docp.kind == lineKindBlockTitle ||
			docp.kind == lineKindID ||
			docp.kind == lineKindIDShort ||
			docp.kind == lineKindText {
			if docp.prevKind == lineKindEmpty {
				break
			}
		}

		listItem.Write(bytes.TrimSpace(line))
		listItem.WriteByte('\n')
		line = nil
	}
	list.postParseList(docp.doc, elKindListDescriptionItem)
	return line
}

// parseListOrdered parser the content as list until it found line that is not
// list-item.
// On success it will return non-empty line and terminator character.
func (docp *documentParser) parseListOrdered(parent *element, title string, line []byte, term int) (got []byte) {
	var (
		logp       = `parseListOrdered`
		itemNumber = 1
		list       = &element{
			kind:     elKindListOrdered,
			rawTitle: title,
		}
		listItem = &element{
			kind:           elKindListOrderedItem,
			listItemNumber: itemNumber,
		}

		el             *element
		parentListItem *element
		ok             bool
	)

	listItem.parseListOrderedItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	line = nil
	for {
		if len(line) == 0 {
			_, line, ok = docp.line(logp)
			if !ok {
				break
			}
		}

		if docp.kind == term {
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = nil
			continue
		}
		if docp.kind == lineKindComment {
			line = nil
			continue
		}
		if docp.kind == lineKindListContinue {
			el, line = docp.parseListBlock()
			if el != nil {
				listItem.addChild(el)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == elKindListOrderedItem {
			el = &element{
				kind: elKindListOrderedItem,
			}

			el.parseListOrderedItem(line)
			if listItem.level == el.level {
				itemNumber++
				el.listItemNumber = itemNumber
				list.addChild(el)
				listItem = el
				line = nil
				continue
			}

			// Case:
			// ... Parent
			// . child
			// ... Next list
			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind && parentListItem.level == el.level {
					list.postParseList(docp.doc, elKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, ``, line, term)
			continue
		}
		if docp.kind == elKindListUnorderedItem {
			el = &element{
				kind: elKindListUnorderedItem,
			}
			el.parseListUnorderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {

					list.postParseList(docp.doc, elKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListUnordered(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindListDescriptionItem {
			el = &element{
				kind: docp.kind,
			}
			el.parseListDescriptionItem(line)

			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {

					list.postParseList(docp.doc, elKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListDescription(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindLiteralParagraph {
			if docp.prevKind == lineKindEmpty {
				el = &element{
					elementAttribute: elementAttribute{
						roles: []string{classNameLiteralBlock},
					},
					kind: docp.kind,
				}
				el.Write(bytes.TrimLeft(line, " \t"))
				el.WriteByte('\n')
				line = docp.consumeLinesUntil(
					el,
					lineKindEmpty,
					[]int{
						lineKindListContinue,
						elKindListOrderedItem,
						elKindListUnorderedItem,
					})
				el.raw = applySubstitutions(docp.doc, el.raw)
				listItem.addChild(el)
				continue
			}
		}
		if docp.kind == elKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockListing ||
			docp.kind == elKindBlockExample ||
			docp.kind == elKindBlockSidebar {
			break
		}
		if docp.kind == elKindSectionL1 ||
			docp.kind == elKindSectionL2 ||
			docp.kind == elKindSectionL3 ||
			docp.kind == elKindSectionL4 ||
			docp.kind == elKindSectionL5 ||
			docp.kind == lineKindAdmonition ||
			docp.kind == lineKindAttributeElement ||
			docp.kind == lineKindBlockTitle ||
			docp.kind == lineKindID ||
			docp.kind == lineKindIDShort ||
			docp.kind == lineKindText {
			if docp.prevKind == lineKindEmpty {
				break
			}
		}

		listItem.Write(bytes.TrimSpace(line))
		listItem.WriteByte('\n')
		line = nil
	}
	list.postParseList(docp.doc, elKindListOrderedItem)
	return line
}

func (docp *documentParser) parseListUnordered(parent, el *element, line []byte, term int) (got []byte) {
	var (
		logp = `parseListUnordered`
		list = &element{
			elementAttribute: elementAttribute{
				roles: []string{classNameUlist},
			},
			kind:     elKindListUnordered,
			rawTitle: el.rawTitle,
		}
		elAttr = &elementAttribute{}

		listItem       *element
		parentListItem *element
		role           string
		ok             bool
	)

	if len(el.rawStyle) > 0 {
		list.addRole(el.rawStyle)
		list.rawStyle = el.rawStyle
	}
	for _, role = range el.roles {
		list.addRole(role)
	}

	listItem = &element{
		kind: elKindListUnorderedItem,
	}
	listItem.parseListUnorderedItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	for _, role = range listItem.roles {
		list.addRole(role)
		if role == classNameChecklist {
			list.rawStyle = role
		}
	}
	parent.addChild(list)

	line = nil
	for {
		if len(line) == 0 {
			_, line, ok = docp.line(logp)
			if !ok {
				break
			}
		}

		if docp.kind == term {
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = nil
			continue
		}
		if docp.kind == lineKindComment {
			line = nil
			continue
		}
		if docp.kind == lineKindListContinue {
			el, line = docp.parseListBlock()
			if el != nil {
				listItem.addChild(el)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == elKindListOrderedItem {
			el = &element{
				kind: elKindListOrderedItem,
			}
			el.parseListOrderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc,
						elKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, ``, line, term)
			continue
		}

		if docp.kind == elKindListUnorderedItem {
			el = &element{
				kind: elKindListUnorderedItem,
			}
			if len(elAttr.rawStyle) > 0 {
				el.addRole(el.rawStyle)
				el.rawStyle = elAttr.rawStyle
				elAttr = &elementAttribute{}
			}

			el.parseListUnorderedItem(line)
			if listItem.level == el.level {
				list.addChild(el)
				for _, role = range listItem.roles {
					list.addRole(role)
					if role == classNameChecklist {
						list.rawStyle = role
					}
				}
				listItem = el
				line = nil
				continue
			}

			// Case:
			// *** Parent
			// * child
			// *** Next list
			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind && parentListItem.level == el.level {
					list.postParseList(docp.doc, elKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListUnordered(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindListDescriptionItem {
			el = &element{
				kind: docp.kind,
			}
			el.parseListDescriptionItem(line)

			parentListItem = parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc,
						elKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListDescription(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindLiteralParagraph {
			if docp.prevKind == lineKindEmpty {
				el = &element{
					elementAttribute: elementAttribute{
						roles: []string{classNameLiteralBlock},
					},
					kind: docp.kind,
				}
				el.Write(bytes.TrimLeft(line, " \t"))
				el.WriteByte('\n')
				line = docp.consumeLinesUntil(
					el,
					lineKindEmpty,
					[]int{
						lineKindListContinue,
						elKindListOrderedItem,
						elKindListUnorderedItem,
					})
				el.raw = applySubstitutions(docp.doc, el.raw)
				listItem.addChild(el)
				continue
			}
		}
		if docp.kind == elKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			el = &element{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(el,
				lineKindEmpty,
				[]int{
					elKindListOrderedItem,
					elKindListUnorderedItem,
				})
			el.raw = applySubstitutions(docp.doc, el.raw)
			listItem.addChild(el)
			continue
		}
		if docp.kind == elKindBlockListing ||
			docp.kind == elKindBlockExample ||
			docp.kind == elKindBlockSidebar {
			break
		}
		if docp.kind == elKindSectionL1 ||
			docp.kind == elKindSectionL2 ||
			docp.kind == elKindSectionL3 ||
			docp.kind == elKindSectionL4 ||
			docp.kind == elKindSectionL5 ||
			docp.kind == lineKindAdmonition ||
			docp.kind == lineKindBlockTitle ||
			docp.kind == lineKindID ||
			docp.kind == lineKindIDShort ||
			docp.kind == lineKindText {
			if docp.prevKind == lineKindEmpty {
				break
			}
		}
		if docp.kind == lineKindAttributeElement {
			if docp.prevKind == lineKindEmpty {
				break
			}
			// Case:
			// * item 1
			// [circle] <-- we are here.
			// ** item 2
			elAttr.parseElementAttribute(line)
			line = nil
			continue
		}

		listItem.Write(bytes.TrimSpace(line))
		listItem.WriteByte('\n')
		line = nil
	}
	list.postParseList(docp.doc, elKindListUnorderedItem)
	return line
}

func (docp *documentParser) parseParagraph(parent, el *element, line []byte, term int) []byte {
	el.kind = elKindParagraph
	el.Write(line)
	el.WriteByte('\n')
	line = docp.consumeLinesUntil(
		el,
		lineKindEmpty,
		[]int{
			term,
			elKindBlockListing,
			elKindBlockListingNamed,
			elKindBlockLiteral,
			elKindBlockLiteralNamed,
			lineKindListContinue,
		})
	el.postParseParagraph(parent)
	el.parseInlineMarkup(docp.doc, elKindText)
	return line
}

func (docp *documentParser) skipCommentAndEmptyLine() (line []byte, ok bool) {
	var logp = `skipCommentAndEmptyLine`

	for {
		_, line, ok = docp.line(logp)
		if !ok {
			return nil, false
		}
		if docp.kind == lineKindEmpty {
			continue
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if docp.kind == lineKindComment {
			continue
		}
		break
	}
	return line, true
}
