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

func parse(doc *Document, content []byte) {
	docp := &documentParser{
		doc: doc,
	}

	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	docp.lines = bytes.Split(content, []byte("\n"))

	docp.parseHeader()
	docp.doc.postParseHeader()

	sectLevel, ok := doc.Attributes[metaNameSectNumLevel]
	if ok {
		doc.sectLevel, _ = strconv.Atoi(sectLevel)
	}

	preamble := &adocNode{
		elementAttribute: elementAttribute{
			Attrs: make(map[string]string),
		},
		kind: nodeKindPreamble,
	}
	doc.content.addChild(preamble)

	docp.parseBlock(preamble, 0)
}

func parseSub(parentDoc *Document, content []byte) (subdoc *Document) {
	subdoc = newDocument()

	for k, v := range parentDoc.Attributes {
		subdoc.Attributes[k] = v
	}

	docp := &documentParser{
		doc:   subdoc,
		lines: bytes.Split(content, []byte("\n")),
	}

	docp.parseBlock(subdoc.content, 0)

	return subdoc
}

func (docp *documentParser) consumeLinesUntil(
	node *adocNode, term int, terms []int,
) (
	line []byte,
) {
	var (
		ok           bool
		allowComment bool
		spaces       []byte
	)
	if term == nodeKindBlockListing || term == nodeKindBlockListingNamed ||
		term == nodeKindLiteralParagraph {
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
				node.Write(line)
				node.WriteByte('\n')
			}
			continue
		}
		if docp.kind == term {
			node.raw = bytes.TrimRight(node.raw, " \n")
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
				node.Write(line)
				node.WriteByte('\n')
				line = nil
				continue
			}
			// Include the content of file into the current
			// document.
			docp.include(elInclude)
			line = nil
			continue
		}
		if node.kind == nodeKindBlockPassthrough ||
			node.kind == nodeKindBlockListing ||
			node.kind == nodeKindBlockLiteral {
			if node.kind != nodeKindTable {
				node.Write(spaces)
			}
		} else if node.kind == nodeKindParagraph && len(spaces) > 0 {
			node.WriteByte(' ')
		}
		node.Write(line)
		node.WriteByte('\n')
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

func (docp *documentParser) parseBlock(parent *adocNode, term int) {
	node := &adocNode{
		kind: nodeKindUnknown,
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
			node.kind = docp.kind
			parent.addChild(node)
			node = &adocNode{}
			line = nil
			continue

		case lineKindID:
			idLabel := line[2 : len(line)-2]
			id, label := parseIDLabel(idLabel)
			if len(id) > 0 {
				node.ID = docp.doc.registerAnchor(
					string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindIDShort:
			id := line[2 : len(line)-1]
			id, label := parseIDLabel(id)
			if len(id) > 0 {
				node.ID = docp.doc.registerAnchor(
					string(id), string(label))
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindInclude:
			elInclude := parseInclude(docp.doc, []byte(line))
			if elInclude == nil {
				node.Write(line)
				node.WriteByte('\n')
				line = nil
				continue
			}
			// Include the content of file into the current
			// document.
			docp.include(elInclude)
			line = nil
			continue

		case lineKindPageBreak:
			node.kind = docp.kind
			parent.addChild(node)
			node = &adocNode{}
			line = nil
			continue

		case lineKindAttribute:
			key, value, ok := parseAttribute(line, false)
			if ok {
				if key == attrNameIcons {
					if node.Attrs == nil {
						node.Attrs = make(map[string]string)
					}
					node.Attrs[key] = value
				} else {
					docp.doc.Attributes.apply(key, value)
					parent.addChild(&adocNode{
						kind:  docp.kind,
						key:   key,
						value: value,
					})
				}
				line = nil
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindAttributeElement:
			node.parseElementAttribute(line)
			if node.style > 0 {
				if isStyleAdmonition(node.style) {
					node.setStyleAdmonition(node.rawStyle)
				}
			}
			line = nil
			continue

		case lineKindStyleClass:
			node.parseStyleClass(line)
			line = nil
			continue

		case lineKindText, lineKindListContinue:
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindBlockTitle:
			node.rawTitle = string(line[1:])
			line = nil
			continue

		case lineKindAdmonition:
			node.kind = nodeKindParagraph
			node.style |= styleAdmonition
			node.parseLineAdmonition(line)
			line = docp.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					term,
					nodeKindBlockListing,
					nodeKindBlockListingNamed,
					nodeKindBlockLiteral,
					nodeKindBlockLiteralNamed,
					lineKindListContinue,
				})
			node.parseInlineMarkup(docp.doc, nodeKindText)
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindSectionL1, nodeKindSectionL2,
			nodeKindSectionL3, nodeKindSectionL4,
			nodeKindSectionL5:
			if term == nodeKindBlockOpen {
				line = docp.parseParagraph(parent, node, line, term)
				parent.addChild(node)
				node = new(adocNode)
				continue
			}

			node.kind = docp.kind
			// BUG: "= =a" could become "a", it should be "=a"
			node.Write(bytes.TrimLeft(line, "= \t"))

			isDiscrete := node.style&styleSectionDiscrete > 0
			if isDiscrete {
				node.kind = nodeKindSectionDiscrete
				node.level = docp.kind
				node.parseSection(docp.doc, isDiscrete)
				parent.addChild(node)
				node = new(adocNode)
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
			node.parseSection(docp.doc, false)
			parent.addChild(node)
			parent = node
			node = new(adocNode)
			line = nil
			continue

		case nodeKindLiteralParagraph:
			if node.isStyleAdmonition() {
				line = docp.parseParagraph(parent, node,
					line, term)
			} else {
				node.kind = docp.kind
				node.addRole(classNameLiteralBlock)
				node.Write(line)
				node.WriteByte('\n')
				line = docp.consumeLinesUntil(
					node,
					lineKindEmpty,
					[]int{
						term,
						nodeKindBlockListing,
						nodeKindBlockListingNamed,
						nodeKindBlockLiteral,
						nodeKindBlockLiteralNamed,
					})
				node.raw = applySubstitutions(docp.doc, node.raw)
			}
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockLiteral:
			node.kind = docp.kind
			node.addRole(classNameLiteralBlock)
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockLiteralNamed:
			node.kind = docp.kind
			node.addRole(classNameLiteralBlock)
			line = docp.consumeLinesUntil(node, lineKindEmpty, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListing:
			node.kind = docp.kind
			node.addRole(classNameListingBlock)
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListingNamed:
			node.kind = docp.kind
			node.addRole(classNameListingBlock)
			line = docp.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					nodeKindBlockListing,
					nodeKindBlockListingNamed,
					nodeKindBlockLiteral,
					nodeKindBlockLiteralNamed,
					lineKindListContinue,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockPassthrough:
			node.kind = docp.kind
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListOrderedItem:
			line = docp.parseListOrdered(parent, node.rawTitle, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListUnorderedItem:
			line = docp.parseListUnordered(parent, node, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListDescriptionItem:
			line = docp.parseListDescription(parent, node, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockImage:
			lineImage := bytes.TrimRight(line[7:], " \t")
			if node.parseBlockImage(docp.doc, lineImage) {
				node.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, node, line, term)
			}
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockOpen, nodeKindBlockExample, nodeKindBlockSidebar:
			node.kind = docp.kind
			docp.parseBlock(node, docp.kind)
			parent.addChild(node)
			node = new(adocNode)
			line = nil
			continue

		case nodeKindBlockExcerpts:
			node.kind = docp.kind
			if node.isStyleVerse() {
				line = docp.consumeLinesUntil(
					node,
					docp.kind,
					[]int{
						term,
						nodeKindBlockListing,
						nodeKindBlockListingNamed,
						nodeKindBlockLiteral,
						nodeKindBlockLiteralNamed,
						lineKindListContinue,
					})
			} else {
				docp.parseBlock(node, docp.kind)
				line = nil
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockVideo:
			if node.parseBlockVideo(docp.doc, line) {
				node.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, node, line, term)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockAudio:
			if node.parseBlockAudio(docp.doc, line) {
				node.kind = docp.kind
				line = nil
			} else {
				line = docp.parseParagraph(parent, node, line, term)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindMacroTOC:
			node.kind = docp.kind
			parent.addChild(node)
			node = new(adocNode)

		case nodeKindTable:
			node.kind = docp.kind
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			parent.addChild(node)
			node.postConsumeTable()
			node = &adocNode{}
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
func (docp *documentParser) parseListBlock() (node *adocNode, line []byte) {
	var ok bool
	for {
		_, line, ok = docp.line()
		if !ok {
			break
		}

		if docp.kind == lineKindAdmonition {
			node = &adocNode{
				elementAttribute: elementAttribute{
					style: styleAdmonition,
				},
				kind: nodeKindParagraph,
			}
			node.parseLineAdmonition(line)
			line = docp.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					nodeKindBlockListing,
					nodeKindBlockListingNamed,
					nodeKindBlockLiteral,
					nodeKindBlockLiteralNamed,
					lineKindListContinue,
				})
			node.parseInlineMarkup(docp.doc, nodeKindText)
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
			return node, line
		}
		if docp.kind == lineKindListContinue {
			continue
		}
		if docp.kind == nodeKindLiteralParagraph {
			node = &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			node.Write(bytes.TrimLeft(line, " \t"))
			node.WriteByte('\n')
			line = docp.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			break
		}
		if docp.kind == lineKindText {
			node = &adocNode{
				kind: nodeKindParagraph,
			}
			node.Write(line)
			node.WriteByte('\n')
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
					nodeKindListDescriptionItem,
				})
			node.postParseParagraph(nil)
			node.parseInlineMarkup(docp.doc, nodeKindText)
			break
		}
		if docp.kind == nodeKindBlockListing {
			node = &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			line = nil
			break
		}
		if docp.kind == nodeKindListOrderedItem {
			break
		}
		if docp.kind == nodeKindListUnorderedItem {
			break
		}
		if docp.kind == nodeKindListDescriptionItem {
			break
		}
	}
	return node, line
}

func (docp *documentParser) parseListDescription(
	parent, node *adocNode, line []byte,
) (
	got []byte,
) {
	list := &adocNode{
		elementAttribute: elementAttribute{
			style: node.style,
		},
		kind:     nodeKindListDescription,
		rawTitle: node.rawTitle,
	}
	listItem := &adocNode{
		elementAttribute: elementAttribute{
			style: list.style,
		},
		kind: nodeKindListDescriptionItem,
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
			var node *adocNode
			node, line = docp.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == nodeKindListOrderedItem {
			line = docp.parseListOrdered(listItem, "", line)
			continue
		}
		if docp.kind == nodeKindListUnorderedItem {
			line = docp.parseListUnordered(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindListDescriptionItem {
			node := &adocNode{
				elementAttribute: elementAttribute{
					style: list.style,
				},
				kind: nodeKindListDescriptionItem,
			}
			node.parseListDescriptionItem(line)
			if listItem.level == node.level {
				list.addChild(node)
				listItem = node
				line = nil
				continue
			}
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == node.level {
					list.postParseList(docp.doc,
						nodeKindListDescriptionItem)
					return line
				}
				parentListItem = parentListItem.parent
			}
			line = docp.parseListDescription(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockListing ||
			docp.kind == nodeKindBlockExample ||
			docp.kind == nodeKindBlockSidebar {
			break
		}
		if docp.kind == nodeKindSectionL1 ||
			docp.kind == nodeKindSectionL2 ||
			docp.kind == nodeKindSectionL3 ||
			docp.kind == nodeKindSectionL4 ||
			docp.kind == nodeKindSectionL5 ||
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
	list.postParseList(docp.doc, nodeKindListDescriptionItem)
	return line
}

//
// parseListOrdered parser the content as list until it found line that is not
// list-item.
// On success it will return non-empty line and terminator character.
//
func (docp *documentParser) parseListOrdered(
	parent *adocNode, title string, line []byte,
) (
	got []byte,
) {
	list := &adocNode{
		kind:     nodeKindListOrdered,
		rawTitle: title,
	}
	listItem := &adocNode{
		kind: nodeKindListOrderedItem,
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
			var node *adocNode
			node, line = docp.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == nodeKindListOrderedItem {
			node := &adocNode{
				kind: nodeKindListOrderedItem,
			}
			node.parseListOrderedItem(line)
			if listItem.level == node.level {
				list.addChild(node)
				listItem = node
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
					parentListItem.level == node.level {
					list.postParseList(docp.doc, nodeKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, "", line)
			continue
		}
		if docp.kind == nodeKindListUnorderedItem {
			node := &adocNode{
				kind: nodeKindListUnorderedItem,
			}
			node.parseListUnorderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == node.level {

					list.postParseList(docp.doc, nodeKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListUnordered(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindListDescriptionItem {
			node := &adocNode{
				kind: docp.kind,
			}
			node.parseListDescriptionItem(line)

			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == node.level {

					list.postParseList(docp.doc, nodeKindListOrderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListDescription(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindLiteralParagraph {
			if docp.prevKind == lineKindEmpty {
				node := &adocNode{
					elementAttribute: elementAttribute{
						roles: []string{classNameLiteralBlock},
					},
					kind: docp.kind,
				}
				node.Write(bytes.TrimLeft(line, " \t"))
				node.WriteByte('\n')
				line = docp.consumeLinesUntil(
					node,
					lineKindEmpty,
					[]int{
						lineKindListContinue,
						nodeKindListOrderedItem,
						nodeKindListUnorderedItem,
					})
				node.raw = applySubstitutions(docp.doc, node.raw)
				listItem.addChild(node)
				continue
			}
		}
		if docp.kind == nodeKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockListing ||
			docp.kind == nodeKindBlockExample ||
			docp.kind == nodeKindBlockSidebar {
			break
		}
		if docp.kind == nodeKindSectionL1 ||
			docp.kind == nodeKindSectionL2 ||
			docp.kind == nodeKindSectionL3 ||
			docp.kind == nodeKindSectionL4 ||
			docp.kind == nodeKindSectionL5 ||
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
	list.postParseList(docp.doc, nodeKindListOrderedItem)
	return line
}

func (docp *documentParser) parseListUnordered(
	parent, node *adocNode, line []byte,
) (
	got []byte,
) {
	list := &adocNode{
		elementAttribute: elementAttribute{
			roles: []string{classNameUlist},
		},
		kind:     nodeKindListUnordered,
		rawTitle: node.rawTitle,
	}
	if len(node.rawStyle) > 0 {
		list.addRole(node.rawStyle)
		list.rawStyle = node.rawStyle
	}
	for _, role := range node.roles {
		list.addRole(role)
	}
	listItem := &adocNode{
		kind: nodeKindListUnorderedItem,
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
			var node *adocNode
			node, line = docp.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if docp.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if docp.kind == nodeKindListOrderedItem {
			node := &adocNode{
				kind: nodeKindListOrderedItem,
			}
			node.parseListOrderedItem(line)

			// Case:
			// . Parent
			// * child
			// . Next list
			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == node.level {
					list.postParseList(docp.doc,
						nodeKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListOrdered(listItem, "", line)
			continue
		}

		if docp.kind == nodeKindListUnorderedItem {
			node := &adocNode{
				kind: nodeKindListUnorderedItem,
			}
			node.parseListUnorderedItem(line)
			if listItem.level == node.level {
				list.addChild(node)
				for _, role := range listItem.roles {
					list.addRole(role)
					if role == classNameChecklist {
						list.rawStyle = role
					}
				}
				listItem = node
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
					parentListItem.level == node.level {
					list.postParseList(docp.doc,
						nodeKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListUnordered(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindListDescriptionItem {
			node := &adocNode{
				kind: docp.kind,
			}
			node.parseListDescriptionItem(line)

			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == docp.kind &&
					parentListItem.level == node.level {
					list.postParseList(docp.doc,
						nodeKindListUnorderedItem)
					return line
				}
				parentListItem = parentListItem.parent
			}

			line = docp.parseListDescription(listItem, node, line)
			continue
		}
		if docp.kind == nodeKindLiteralParagraph {
			if docp.prevKind == lineKindEmpty {
				node = &adocNode{
					elementAttribute: elementAttribute{
						roles: []string{classNameLiteralBlock},
					},
					kind: docp.kind,
				}
				node.Write(bytes.TrimLeft(line, " \t"))
				node.WriteByte('\n')
				line = docp.consumeLinesUntil(
					node,
					lineKindEmpty,
					[]int{
						lineKindListContinue,
						nodeKindListOrderedItem,
						nodeKindListUnorderedItem,
					})
				node.raw = applySubstitutions(docp.doc, node.raw)
				listItem.addChild(node)
				continue
			}
		}
		if docp.kind == nodeKindBlockListingNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameListingBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockLiteralNamed {
			if docp.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				elementAttribute: elementAttribute{
					roles: []string{classNameLiteralBlock},
				},
				kind: docp.kind,
			}
			line = docp.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.raw = applySubstitutions(docp.doc, node.raw)
			listItem.addChild(node)
			continue
		}
		if docp.kind == nodeKindBlockListing ||
			docp.kind == nodeKindBlockExample ||
			docp.kind == nodeKindBlockSidebar {
			break
		}
		if docp.kind == nodeKindSectionL1 ||
			docp.kind == nodeKindSectionL2 ||
			docp.kind == nodeKindSectionL3 ||
			docp.kind == nodeKindSectionL4 ||
			docp.kind == nodeKindSectionL5 ||
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
	list.postParseList(docp.doc, nodeKindListUnorderedItem)
	return line
}

func (docp *documentParser) parseParagraph(
	parent, node *adocNode, line []byte, term int,
) []byte {
	node.kind = nodeKindParagraph
	node.Write(line)
	node.WriteByte('\n')
	line = docp.consumeLinesUntil(
		node,
		lineKindEmpty,
		[]int{
			term,
			nodeKindBlockListing,
			nodeKindBlockListingNamed,
			nodeKindBlockLiteral,
			nodeKindBlockLiteralNamed,
			lineKindListContinue,
		})
	node.postParseParagraph(parent)
	node.parseInlineMarkup(docp.doc, nodeKindText)
	return line
}
