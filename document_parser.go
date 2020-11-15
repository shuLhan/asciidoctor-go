// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/parser"
)

type documentParser struct {
	doc      *Document
	p        *parser.Parser
	lineNum  int
	prevKind int
	kind     int
}

//
// Parse the content into a Document.
//
func Parse(content []byte) (doc *Document) {
	doc = newDocument()

	docp := &documentParser{
		doc: doc,
		p:   parser.New(string(content), "\n"),
	}

	docp.parseHeader()
	docp.doc.postParseHeader()

	sectLevel, ok := doc.Attributes[attrNameSectnumlevels]
	if ok {
		doc.sectLevel, _ = strconv.Atoi(sectLevel)
	}

	preamble := &adocNode{
		kind:  nodeKindPreamble,
		Attrs: make(map[string]string),
	}
	doc.content.addChild(preamble)

	docp.parseBlock(preamble, 0)

	return doc
}

func (docp *documentParser) consumeLinesUntil(
	node *adocNode, term int, terms []int,
) (
	line string,
) {
	var (
		c      rune
		spaces string
	)
	for {
		spaces, line, c = docp.line()
		if len(line) == 0 && c == 0 {
			break
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if docp.kind == lineKindComment {
			continue
		}
		if docp.kind == term {
			node.raw = bytes.TrimRight(node.raw, " \n")
			return ""
		}
		for _, t := range terms {
			if t == docp.kind {
				return line
			}
		}
		if node.kind == nodeKindBlockPassthrough ||
			node.kind == nodeKindBlockListing ||
			node.kind == nodeKindBlockLiteral {
			node.WriteString(spaces)
		} else if node.kind == nodeKindParagraph && len(spaces) > 0 {
			node.WriteByte(' ')
		}
		node.WriteString(line)
		node.WriteByte('\n')
	}
	return line
}

func (docp *documentParser) line() (spaces, line string, c rune) {
	docp.prevKind = docp.kind
	line, c = docp.p.Line()
	if len(line) > 0 || c > 0 {
		docp.lineNum++
	}
	docp.kind, spaces, line = whatKindOfLine(line)
	if debug.Value >= 1 {
		fmt.Printf("line %3d: kind %3d: c %3d: %s\n", docp.lineNum,
			docp.kind, c, line)
	}
	return spaces, line, c
}

func (docp *documentParser) parseBlock(parent *adocNode, term int) {
	node := &adocNode{
		kind: nodeKindUnknown,
	}
	var (
		spaces, line string
		c            rune
	)
	for {
		if len(line) == 0 {
			spaces, line, c = docp.line()
			if len(line) == 0 && c == 0 {
				return
			}
		}

		switch docp.kind {
		case term:
			return
		case lineKindEmpty:
			line = ""
			continue
		case lineKindBlockComment:
			docp.parseIgnoreCommentBlock()
			line = ""
			continue
		case lineKindComment:
			line = ""
			continue
		case lineKindHorizontalRule:
			node.kind = docp.kind
			parent.addChild(node)
			node = &adocNode{}
			line = ""
			continue

		case lineKindID:
			idLabel := line[2 : len(line)-2]
			id, label := parseIDLabel(idLabel)
			if len(id) > 0 {
				node.ID = docp.doc.registerAnchor(id, label)
				line = ""
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
				node.ID = docp.doc.registerAnchor(id, label)
				line = ""
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindPageBreak:
			node.kind = docp.kind
			parent.addChild(node)
			node = &adocNode{}
			line = ""
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
				line = ""
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindAttributeElement:
			key, val, opts := parseAttributeElement(line)

			styleKind := parseStyle(key)
			if styleKind > 0 {
				node.style |= styleKind
				if isStyleAdmonition(styleKind) {
					node.setStyleAdmonition(key)
				} else if isStyleQuote(styleKind) {
					node.setQuoteOpts(opts[1:])
				} else if isStyleVerse(styleKind) {
					node.setQuoteOpts(opts[1:])
				}
				line = ""
				continue
			}
			if key == attrNameRefText {
				if node.Attrs == nil {
					node.Attrs = make(map[string]string)
				}
				node.Attrs[key] = val
				line = ""
				continue
			}
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindStyleClass:
			node.parseStyleClass(line)
			line = ""
			continue

		case lineKindText, lineKindListContinue:
			line = docp.parseParagraph(parent, node, line, term)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindBlockTitle:
			node.rawTitle = line[1:]
			line = ""
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
			node.WriteString(
				// BUG: "= =a" could become "a", it should be "=a"
				strings.TrimLeft(line, "= \t"),
			)

			var expParent = docp.kind - 1
			for parent.kind != expParent {
				parent = parent.parent
				if parent == nil {
					parent = docp.doc.content
					break
				}
			}
			node.parseSection(docp.doc)
			parent.addChild(node)
			parent = node
			node = new(adocNode)
			line = ""
			continue

		case nodeKindLiteralParagraph:
			if node.IsStyleAdmonition() {
				line = docp.parseParagraph(parent, node,
					spaces+line, term)
			} else {
				node.kind = docp.kind
				node.classes = append(node.classes, classNameLiteralBlock)
				node.WriteString(line)
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
			node.classes = append(node.classes, classNameLiteralBlock)
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockLiteralNamed:
			node.kind = docp.kind
			node.classes = append(node.classes, classNameLiteralBlock)
			line = docp.consumeLinesUntil(node, lineKindEmpty, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListing:
			node.kind = docp.kind
			node.classes = append(node.classes, classNameListingBlock)
			line = docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListingNamed:
			node.kind = docp.kind
			node.classes = append(node.classes, classNameListingBlock)
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
			lineImage := strings.TrimRight(line[7:], " \t")
			if node.parseBlockImage(docp.doc, lineImage) {
				node.kind = docp.kind
				line = ""
			} else {
				line = docp.parseParagraph(parent, node, line, term)
			}
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockOpen, nodeKindBlockExample,
			nodeKindBlockSidebar:
			node.kind = docp.kind
			docp.parseBlock(node, docp.kind)
			parent.addChild(node)
			node = new(adocNode)
			line = ""
			continue

		case nodeKindBlockExcerpts:
			node.kind = docp.kind
			if node.IsStyleVerse() {
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
				line = ""
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockVideo:
			if node.parseBlockVideo(docp.doc, line) {
				node.kind = docp.kind
				line = ""
			} else {
				line = docp.parseParagraph(parent, node,
					"video::"+line, term)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockAudio:
			if node.parseBlockAudio(docp.doc, line) {
				node.kind = docp.kind
				line = ""
			} else {
				line = docp.parseParagraph(parent, node,
					"audio::"+line, term)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindMacroTOC:
			node.kind = docp.kind
			parent.addChild(node)
			node = new(adocNode)
		}
		line = ""
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
		_, line, c := docp.line()
		if len(line) == 0 && c == 0 {
			return
		}
		if len(line) == 0 {
			return
		}
		if strings.HasPrefix(line, "////") {
			docp.parseIgnoreCommentBlock()
			continue
		}
		if strings.HasPrefix(line, "//") {
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
				docp.doc.header.WriteString(strings.TrimSpace(line[2:]))
				docp.doc.Title.raw = string(docp.doc.header.raw)
				state = stateAuthor
			} else {
				docp.doc.Author = line
				state = stateRevision
			}
			continue
		}
		switch state {
		case stateAuthor:
			docp.doc.Author = line
			state = stateRevision

		case stateRevision:
			if !docp.parseHeaderRevision(line) {
				return
			}
			state = stateEnd
		}
	}
}

//
//	DOC_REVISION     = DOC_REV_VERSION [ "," DOC_REV_DATE ]
//
//	DOC_REV_VERSION  = "v" 1*DIGIT "." 1*DIGIT "." 1*DIGIT
//
//	DOC_REV_DATE     = 1*2DIGIT WSP 3*ALPHA WSP 4*DIGIT
//
func (docp *documentParser) parseHeaderRevision(line string) bool {
	if line[0] != 'v' {
		return false
	}

	idx := strings.IndexByte(line, ',')
	if idx > 0 {
		docp.doc.RevNumber = line[1:idx]
		docp.doc.RevDate = strings.TrimSpace(line[idx+1:])
		docp.doc.RevSeparator = ","
	} else {
		docp.doc.RevNumber = line[1:]
	}
	return true
}

func (docp *documentParser) parseIgnoreCommentBlock() {
	for {
		line, c := docp.p.Line()
		if len(line) == 0 && c == 0 {
			return
		}
		if strings.HasPrefix(line, "////") {
			return
		}
	}
}

//
// parseListBlock parse block after list continuation "+" until we found
// empty line or non-list line.
//
func (docp *documentParser) parseListBlock() (node *adocNode, line string) {
	var c rune
	for {
		_, line, c = docp.line()
		if len(line) == 0 && c == 0 {
			break
		}

		if docp.kind == lineKindAdmonition {
			node = &adocNode{
				kind:  nodeKindParagraph,
				style: styleAdmonition,
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
				kind:    docp.kind,
				classes: []string{classNameLiteralBlock},
			}
			node.WriteString(strings.TrimLeft(line, " \t"))
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
			node.WriteString(line)
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
				kind:    docp.kind,
				classes: []string{classNameListingBlock},
			}
			docp.consumeLinesUntil(node, docp.kind, nil)
			node.raw = applySubstitutions(docp.doc, node.raw)
			line = ""
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

func (docp *documentParser) parseListDescription(parent, node *adocNode, line string) (
	got string,
) {
	list := &adocNode{
		kind:     nodeKindListDescription,
		rawTitle: node.rawTitle,
		style:    node.style,
	}
	listItem := &adocNode{
		kind:  nodeKindListDescriptionItem,
		style: list.style,
	}
	listItem.parseListDescriptionItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	var (
		c rune
	)
	line = ""
	for {
		if len(line) == 0 {
			_, line, c = docp.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}
		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if docp.kind == lineKindComment {
			line = ""
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
				kind:  nodeKindListDescriptionItem,
				style: list.style,
			}
			node.parseListDescriptionItem(line)
			if listItem.level == node.level {
				list.addChild(node)
				listItem = node
				line = ""
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
				kind:    docp.kind,
				classes: []string{classNameListingBlock},
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
				kind:    docp.kind,
				classes: []string{classNameLiteralBlock},
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

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(docp.doc, nodeKindListDescriptionItem)
	return line
}

//
// parseListOrdered parser the content as list until it found line that is not
// list-item.
// On success it will return non-empty line and terminator character.
//
func (docp *documentParser) parseListOrdered(parent *adocNode, title, line string) (
	got string,
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

	var c rune
	line = ""
	for {
		if len(line) == 0 {
			_, line, c = docp.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}

		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if docp.kind == lineKindComment {
			line = ""
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
				line = ""
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
					kind:    docp.kind,
					classes: []string{classNameLiteralBlock},
				}
				node.WriteString(strings.TrimLeft(line, " \t"))
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
				kind:    docp.kind,
				classes: []string{classNameListingBlock},
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
				kind:    docp.kind,
				classes: []string{classNameLiteralBlock},
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

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(docp.doc, nodeKindListOrderedItem)
	return line
}

func (docp *documentParser) parseListUnordered(parent, node *adocNode, line string) (
	got string,
) {
	list := &adocNode{
		kind:     nodeKindListUnordered,
		classes:  []string{classNameUlist},
		rawTitle: node.rawTitle,
	}
	listItem := &adocNode{
		kind: nodeKindListUnorderedItem,
	}
	listItem.parseListUnorderedItem(line)
	list.level = listItem.level
	list.addChild(listItem)
	parent.addChild(list)

	var c rune
	line = ""
	for {
		if len(line) == 0 {
			_, line, c = docp.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}

		if docp.kind == lineKindBlockComment {
			docp.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if docp.kind == lineKindComment {
			line = ""
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
				listItem = node
				line = ""
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
					kind:    docp.kind,
					classes: []string{classNameLiteralBlock},
				}
				node.WriteString(strings.TrimLeft(line, " \t"))
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
				kind:    docp.kind,
				classes: []string{classNameListingBlock},
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
				kind:    docp.kind,
				classes: []string{classNameLiteralBlock},
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

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(docp.doc, nodeKindListUnorderedItem)
	return line
}

func (docp *documentParser) parseParagraph(
	parent, node *adocNode, line string, term int,
) string {
	node.kind = nodeKindParagraph
	node.WriteString(line)
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
