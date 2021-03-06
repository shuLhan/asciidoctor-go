// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/shuLhan/share/lib/debug"
)

type documentParser struct {
	doc      *Document
	lines    [][]byte
	lineNum  int
	prevKind int
	kind     int
}

//
// Parse the content into a Document.
//
func Parse(content []byte) (doc *Document) {
	doc = newDocument()
	parse(doc, content)
	return doc
}

func newDocumentParser(doc *Document, content []byte) (docp *documentParser) {
	docp = &documentParser{
		doc: doc,
	}

	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	docp.lines = bytes.Split(content, []byte("\n"))

	return docp
}

func parse(doc *Document, content []byte) {
	docp := newDocumentParser(doc, content)

	docp.parseHeader()
	docp.doc.postParseHeader()

	sectLevel, ok := doc.Attributes[metaNameSectNumLevel]
	if ok {
		doc.sectLevel, _ = strconv.Atoi(sectLevel)
	}

	preamble := &element{
		elementAttribute: elementAttribute{
			Attrs: make(map[string]string),
		},
		kind: elKindPreamble,
	}
	doc.content.addChild(preamble)

	docp.parseBlock(preamble, 0)
}

func parseSub(parentDoc *Document, content []byte) (subdoc *Document) {
	subdoc = newDocument()

	for k, v := range parentDoc.Attributes {
		subdoc.Attributes[k] = v
	}

	docp := newDocumentParser(subdoc, content)

	docp.parseBlock(subdoc.content, 0)

	return subdoc
}

func (docp *documentParser) consumeLinesUntil(
	el *element, term int, terms []int,
) (
	line []byte,
) {
	var (
		ok           bool
		allowComment bool
		spaces       []byte
	)
	if term == elKindBlockListing || term == elKindBlockListingNamed ||
		term == elKindLiteralParagraph {
		allowComment = true
	}
	for {
		spaces, line, ok = docp.line()
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
		for _, t := range terms {
			if t == docp.kind {
				return line
			}
		}
		if docp.kind == lineKindInclude {
			elInclude := parseInclude(docp.doc, line)
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

func (docp *documentParser) include(el *elementInclude) {
	content := bytes.ReplaceAll(el.content, []byte("\r\n"), []byte("\n"))
	content = bytes.TrimRight(content, "\n")
	includedLines := bytes.Split(content, []byte("\n"))
	newLines := make([][]byte, 0, len(docp.lines)+len(includedLines))

	// Do not add the "include" directive
	docp.lineNum--
	newLines = append(newLines, docp.lines[:docp.lineNum]...)
	newLines = append(newLines, includedLines...)
	newLines = append(newLines, docp.lines[docp.lineNum+1:]...)
	docp.lines = newLines
}

//
// line return the next line in the content of raw document.
// It will return ok as false if there are no more line.
//
func (docp *documentParser) line() (spaces, line []byte, ok bool) {
	docp.prevKind = docp.kind

	if docp.lineNum >= len(docp.lines) {
		return nil, nil, false
	}

	line = docp.lines[docp.lineNum]
	docp.lineNum++

	docp.kind, spaces, line = whatKindOfLine(line)
	if debug.Value == 2 {
		fmt.Printf("line %3d: kind %3d: %s\n", docp.lineNum, docp.kind, line)
	}
	return spaces, line, true
}

func (docp *documentParser) parseBlock(parent *element, term int) {
	el := &element{
		kind: elKindUnknown,
	}
	var (
		line []byte
		ok   bool
	)
	for {
		if len(line) == 0 {
			_, line, ok = docp.line()
			if !ok {
				return
			}
		}

		switch docp.kind {
		case term:
			return
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
			idLabel := line[2 : len(line)-2]
			id, label := parseIDLabel(idLabel)
			if len(id) > 0 {
				el.ID = docp.doc.registerAnchor(
					string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindIDShort:
			id := line[2 : len(line)-1]
			id, label := parseIDLabel(id)
			if len(id) > 0 {
				el.ID = docp.doc.registerAnchor(
					string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, el, line, term)
			parent.addChild(el)
			el = &element{}
			continue

		case lineKindInclude:
			elInclude := parseInclude(docp.doc, []byte(line))
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
			key, value, ok := parseAttribute(line, false)
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

		case elKindSectionL1, elKindSectionL2,
			elKindSectionL3, elKindSectionL4,
			elKindSectionL5:
			if term == elKindBlockOpen {
				line = docp.parseParagraph(parent, el, line, term)
				parent.addChild(el)
				el = new(element)
				continue
			}

			el.kind = docp.kind
			// BUG: "= =a" could become "a", it should be "=a"
			el.Write(bytes.TrimLeft(line, "= \t"))

			isDiscrete := el.style&styleSectionDiscrete > 0
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
			lineImage := bytes.TrimRight(line[7:], " \t")
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

//
// parseHeader document header consist of title and optional authors,
// revision, and zero or more attributes.
//
// The document attributes can be in any order, but the author and revision
// MUST be in order.
//
//	DOC_HEADER  = *(DOC_ATTRIBUTE / COMMENTS)
//	              "=" SP *ADOC_WORD LF
//	              (*DOC_ATTRIBUTE)
//	              DOC_AUTHORS LF
//	              (*DOC_ATTRIBUTE)
//	              DOC_REVISION LF
//	              (*DOC_ATTRIBUTE)
//
func (docp *documentParser) parseHeader() {
	const (
		stateTitle int = iota
		stateAuthor
		stateRevision
		stateEnd
	)
	state := stateTitle
	for {
		_, line, ok := docp.line()
		if !ok {
			return
		}
		if len(line) == 0 {
			return
		}
		if bytes.HasPrefix(line, []byte("////")) {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if bytes.HasPrefix(line, []byte("//")) {
			continue
		}
		if line[0] == ':' {
			key, value, ok := parseAttribute(line, false)
			if ok {
				docp.doc.Attributes.apply(key, value)
			}
			continue
		}
		if state == stateTitle {
			if isTitle(line) {
				docp.doc.header.Write(bytes.TrimSpace(line[2:]))
				docp.doc.Title.raw = string(docp.doc.header.raw)
				state = stateAuthor
			} else {
				docp.doc.rawAuthors = string(line)
				state = stateRevision
			}
			continue
		}
		switch state {
		case stateAuthor:
			docp.doc.rawAuthors = string(line)
			state = stateRevision

		case stateRevision:
			docp.doc.rawRevision = string(line)
			state = stateEnd
		}
	}
}

func (docp *documentParser) parseIgnoreCommentBlock() {
	for {
		_, line, ok := docp.line()
		if !ok {
			return
		}
		if bytes.HasPrefix(line, []byte("////")) {
			return
		}
	}
}

//
// parseListBlock parse block after list continuation "+" until we found
// empty line or non-list line.
//
func (docp *documentParser) parseListBlock() (el *element, line []byte) {
	var ok bool
	for {
		_, line, ok = docp.line()
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
					elKindListOrderedItem,
					elKindListUnorderedItem,
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

func (docp *documentParser) parseListDescription(
	parent, el *element, line []byte, term int,
) (
	got []byte,
) {
	list := &element{
		elementAttribute: elementAttribute{
			style: el.style,
		},
		kind:     elKindListDescription,
		rawTitle: el.rawTitle,
	}
	listItem := &element{
		elementAttribute: elementAttribute{
			style: list.style,
		},
		kind: elKindListDescriptionItem,
	}
	listItem.parseListDescriptionItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	line = nil
	var ok bool
	for {
		if len(line) == 0 {
			_, line, ok = docp.line()
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
			var el *element
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
			line = docp.parseListOrdered(listItem, "", line, term)
			continue
		}
		if docp.kind == elKindListUnorderedItem {
			line = docp.parseListUnordered(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindListDescriptionItem {
			el := &element{
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
			parentListItem := parent
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
			el := &element{
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
			el := &element{
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

//
// parseListOrdered parser the content as list until it found line that is not
// list-item.
// On success it will return non-empty line and terminator character.
//
func (docp *documentParser) parseListOrdered(
	parent *element, title string, line []byte, term int,
) (
	got []byte,
) {
	list := &element{
		kind:     elKindListOrdered,
		rawTitle: title,
	}
	listItem := &element{
		kind: elKindListOrderedItem,
	}
	listItem.parseListOrderedItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	var ok bool
	line = nil
	for {
		if len(line) == 0 {
			_, line, ok = docp.line()
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
			var el *element
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
			el := &element{
				kind: elKindListOrderedItem,
			}
			el.parseListOrderedItem(line)
			if listItem.level == el.level {
				list.addChild(el)
				listItem = el
				line = nil
				continue
			}

			// Case:
			// ... Parent
			// . child
			// ... Next list
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc, elKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, "", line, term)
			continue
		}
		if docp.kind == elKindListUnorderedItem {
			el := &element{
				kind: elKindListUnorderedItem,
			}
			el.parseListUnorderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem := parent
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
			el := &element{
				kind: docp.kind,
			}
			el.parseListDescriptionItem(line)

			parentListItem := parent
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
				el := &element{
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
			el := &element{
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
			el := &element{
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

func (docp *documentParser) parseListUnordered(
	parent, el *element, line []byte, term int,
) (
	got []byte,
) {
	list := &element{
		elementAttribute: elementAttribute{
			roles: []string{classNameUlist},
		},
		kind:     elKindListUnordered,
		rawTitle: el.rawTitle,
	}
	if len(el.rawStyle) > 0 {
		list.addRole(el.rawStyle)
		list.rawStyle = el.rawStyle
	}
	for _, role := range el.roles {
		list.addRole(role)
	}
	listItem := &element{
		kind: elKindListUnorderedItem,
	}
	listItem.parseListUnorderedItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	for _, role := range listItem.roles {
		list.addRole(role)
		if role == classNameChecklist {
			list.rawStyle = role
		}
	}
	parent.addChild(list)

	var ok bool
	line = nil
	for {
		if len(line) == 0 {
			_, line, ok = docp.line()
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
			var el *element
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
			el := &element{
				kind: elKindListOrderedItem,
			}
			el.parseListOrderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc,
						elKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, "", line, term)
			continue
		}

		if docp.kind == elKindListUnorderedItem {
			el := &element{
				kind: elKindListUnorderedItem,
			}
			el.parseListUnorderedItem(line)
			if listItem.level == el.level {
				list.addChild(el)
				for _, role := range listItem.roles {
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
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == el.level {
					list.postParseList(docp.doc,
						elKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListUnordered(listItem, el, line, term)
			continue
		}
		if docp.kind == elKindListDescriptionItem {
			el := &element{
				kind: docp.kind,
			}
			el.parseListDescriptionItem(line)

			parentListItem := parent
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
			el := &element{
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
			el := &element{
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
	list.postParseList(docp.doc, elKindListUnorderedItem)
	return line
}

func (docp *documentParser) parseParagraph(
	parent, el *element, line []byte, term int,
) []byte {
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
