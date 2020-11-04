// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/parser"
)

const (
	defSectnumlevels = 3
	defTOCLevel      = 2
	defTOCTitle      = "Table of Contents"
)

//
// Document represent content of asciidoc that has been parsed.
//
type Document struct {
	file string
	p    *parser.Parser

	Author       string
	Title        string
	RevNumber    string
	RevSeparator string
	RevDate      string
	LastUpdated  string
	Attributes   AttributeEntry

	classes []string
	lineNum int

	TOCTitle     string
	TOCLevel     int
	tocPosition  string
	tocIsEnabled bool
	tocClasses   []string

	// anchors contains mapping between unique ID and its label.
	anchors map[string]*anchor
	// titleID is the reverse of anchors, it contains mapping of title and
	// its ID.
	titleID map[string]string

	sectnums  *sectionCounters
	sectLevel int

	title   *adocNode
	header  *adocNode
	content *adocNode

	prevKind int
	kind     int
}

//
// Open the ascidoc file and parse it.
//
func Open(file string) (doc *Document, err error) {
	fi, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("ciigo.Open %s: %w", file, err)
	}

	doc = &Document{
		file:        file,
		LastUpdated: fi.ModTime().Round(time.Second).Format("2006-01-02 15:04:05 Z0700"),
		TOCLevel:    defTOCLevel,
		TOCTitle:    defTOCTitle,
		Attributes:  newAttributeEntry(),
		anchors:     make(map[string]*anchor),
		titleID:     make(map[string]string),
		sectnums:    &sectionCounters{},
		sectLevel:   defSectnumlevels,
		content: &adocNode{
			kind: nodeKindDocContent,
		},
	}

	doc.Parse(raw)

	return doc, nil
}

//
// Classes return the HTML class value for document body.
//
func (doc *Document) Classes() string {
	return strings.Join(doc.classes, " ")
}

//
// Parse the content of asciidoc document.
//
func (doc *Document) Parse(content []byte) {
	var err error

	doc.p = parser.New(string(content), "\n")

	doc.parseHeader()

	doc.title = parseInlineMarkup(doc, []byte(doc.Title))

	doc.Title, err = doc.title.toText()
	if err != nil {
		log.Fatalf("Parse: " + err.Error())
	}

	sectLevel, ok := doc.Attributes[attrNameSectnumlevels]
	if ok {
		doc.sectLevel, err = strconv.Atoi(sectLevel)
		if err != nil {
			log.Fatalf("Parse %s %s: %s", attrNameSectnumlevels,
				sectLevel, err)
		}
	}

	preamble := &adocNode{
		kind:  nodeKindPreamble,
		Attrs: make(map[string]string),
	}
	doc.content.addChild(preamble)

	doc.parseBlock(preamble, 0)
}

//
// ToHTML convert the asciidoc document into full HTML document, including
// head and body.
//
func (doc *Document) ToHTML(w io.Writer) (err error) {
	tmpl, err := doc.createHTMLTemplate()
	if err != nil {
		return err
	}

	doc.classes = append(doc.classes, classNameArticle)

	doc.tocPosition, doc.tocIsEnabled = doc.Attributes[metaNameTOC]

	switch doc.tocPosition {
	case metaValueLeft:
		doc.classes = append(doc.classes, classNameToc2, classNameTocLeft)
		doc.tocClasses = append(doc.tocClasses, classNameToc2)
	case metaValueRight:
		doc.classes = append(doc.classes, classNameToc2, classNameTocRight)
		doc.tocClasses = append(doc.tocClasses, classNameToc2)
	default:
		doc.tocClasses = append(doc.tocClasses, classNameToc)
	}

	err = tmpl.ExecuteTemplate(w, "BEGIN", doc)
	if err != nil {
		return err
	}

	err = doc.htmlWriteBody(tmpl, w)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "END", doc)

	return err
}

//
// ToHTMLBody convert the document object into HTML with content of body only.
//
func (doc *Document) ToHTMLBody(w io.Writer) (err error) {
	tmpl, err := doc.createHTMLTemplate()
	if err != nil {
		return fmt.Errorf("ToHTMLBody: %w", err)
	}

	doc.classes = append(doc.classes, classNameArticle)

	doc.tocPosition, doc.tocIsEnabled = doc.Attributes[metaNameTOC]

	switch doc.tocPosition {
	case metaValueLeft:
		doc.classes = append(doc.classes, classNameToc2, classNameTocLeft)
		doc.tocClasses = append(doc.tocClasses, classNameToc2)
	case metaValueRight:
		doc.classes = append(doc.classes, classNameToc2, classNameTocRight)
		doc.tocClasses = append(doc.tocClasses, classNameToc2)
	default:
		doc.tocClasses = append(doc.tocClasses, classNameToc)
	}

	err = doc.htmlWriteBody(tmpl, w)
	if err != nil {
		return err
	}

	return nil
}

func (doc *Document) htmlWriteBody(tmpl *template.Template, w io.Writer) (err error) {
	err = tmpl.ExecuteTemplate(w, "BEGIN_HEADER", doc)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "BEGIN_TITLE", doc)
	if err != nil {
		return err
	}
	err = doc.title.toHTML(doc, tmpl, w, false)
	if err != nil {
		return err
	}
	err = tmpl.ExecuteTemplate(w, "END_TITLE", doc)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "HEADER_DETAILS", doc)
	if err != nil {
		return err
	}

	if doc.tocIsEnabled && (doc.tocPosition == "" ||
		doc.tocPosition == metaValueAuto ||
		doc.tocPosition == metaValueLeft ||
		doc.tocPosition == metaValueRight) {
		err = doc.tocHTML(tmpl, w)
		if err != nil {
			return fmt.Errorf("ToHTML: %w", err)
		}
	}

	err = tmpl.ExecuteTemplate(w, "END_HEADER", doc)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "BEGIN_CONTENT", doc)
	if err != nil {
		return err
	}

	if doc.content.child != nil {
		err = doc.content.child.toHTML(doc, tmpl, w, false)
		if err != nil {
			return err
		}
	}
	if doc.content.next != nil {
		err = doc.content.next.toHTML(doc, tmpl, w, false)
		if err != nil {
			return err
		}
	}
	return nil
}

//
// TocClasses return list of classes for table of contents.
//
func (doc *Document) TocClasses() string {
	return strings.Join(doc.tocClasses, " ")
}

func (doc *Document) consumeLinesUntil(node *adocNode, term int, terms []int) (
	line string, c rune,
) {
	spaces := ""
	for {
		spaces, line, c = doc.line()
		if len(line) == 0 && c == 0 {
			break
		}
		if doc.kind == lineKindBlockComment {
			doc.parseIgnoreCommentBlock()
			continue
		}
		if doc.kind == lineKindComment {
			continue
		}
		if doc.kind == term {
			node.raw = bytes.TrimRight(node.raw, " \n")
			return "", 0
		}
		for _, t := range terms {
			if t == doc.kind {
				return line, c
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
	return line, c
}

func (doc *Document) line() (spaces, line string, c rune) {
	doc.prevKind = doc.kind
	doc.lineNum++
	line, c = doc.p.Line()
	doc.kind, spaces, line = whatKindOfLine(line)
	if debug.Value >= 2 {
		fmt.Printf("line %3d: kind %3d: %s\n", doc.lineNum, doc.kind, line)
	}
	return spaces, line, c
}

//
// parseAttribute parse document attribute and return its key and optional
// value.
//
//	DOC_ATTRIBUTE  = ":" DOC_ATTR_KEY ":" *STRING LF
//
//	DOC_ATTR_KEY   = ( "toc" / "sectanchors" / "sectlinks"
//	               /   "imagesdir" / "data-uri" / *META_KEY ) LF
//
//	META_KEY_CHAR  = (A..Z | a..z | 0..9 | '_')
//
//	META_KEY       = 1META_KEY_CHAR *(META_KEY_CHAR | '-')
//
func (doc *Document) parseAttribute(line string, strict bool) (key, value string) {
	var sb strings.Builder

	if !(ascii.IsAlnum(line[1]) || line[1] == '_') {
		return "", ""
	}

	sb.WriteByte(line[1])
	x := 2
	for ; x < len(line); x++ {
		if line[x] == ':' {
			break
		}
		if ascii.IsAlnum(line[x]) || line[x] == '_' ||
			line[x] == '-' || line[x] == '!' {
			sb.WriteByte(line[x])
			continue
		}
		if strict {
			return "", ""
		}
	}
	if x == len(line) {
		return "", ""
	}

	return sb.String(), strings.TrimSpace(line[x+1:])
}

func (doc *Document) parseBlock(parent *adocNode, term int) {
	node := &adocNode{
		kind: nodeKindUnknown,
	}
	var (
		spaces, line string
		c            rune
	)
	for {
		if len(line) == 0 {
			spaces, line, c = doc.line()
			if len(line) == 0 && c == 0 {
				return
			}
		}

		switch doc.kind {
		case term:
			return
		case lineKindEmpty:
			line = ""
			continue
		case lineKindBlockComment:
			doc.parseIgnoreCommentBlock()
			line = ""
			continue
		case lineKindComment:
			line = ""
			continue
		case lineKindHorizontalRule:
			node.kind = doc.kind
			parent.addChild(node)
			node = &adocNode{}
			line = ""
			continue

		case lineKindID:
			idLabel := line[2 : len(line)-2]
			id, label := parseIDLabel(idLabel)
			if len(id) > 0 {
				node.ID = doc.registerAnchor(id, label)
				line = ""
				continue
			}
			node.kind = nodeKindParagraph
			node.WriteString(line)
			node.WriteByte('\n')
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindIDShort:
			id := line[2 : len(line)-1]
			id, label := parseIDLabel(id)
			if len(id) > 0 {
				node.ID = doc.registerAnchor(id, label)
				line = ""
				continue
			}
			node.kind = nodeKindParagraph
			node.WriteString(line)
			node.WriteByte('\n')
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindPageBreak:
			node.kind = doc.kind
			parent.addChild(node)
			node = &adocNode{}
			line = ""
			continue

		case lineKindAttribute:
			key, value := doc.parseAttribute(line, false)
			if len(key) > 0 {
				if key == attrNameIcons {
					if node.Attrs == nil {
						node.Attrs = make(map[string]string)
					}
					node.Attrs[key] = value
				} else {
					doc.Attributes.apply(key, value)
					parent.addChild(&adocNode{
						kind:  doc.kind,
						key:   key,
						value: value,
					})
				}
				line = ""
				continue
			}
			node.kind = nodeKindParagraph
			node.WriteString(line)
			node.WriteByte('\n')
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
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
			node.kind = nodeKindParagraph
			node.WriteString(line)
			node.WriteByte('\n')
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case lineKindStyleClass:
			node.parseStyleClass(line)
			line = ""
			continue

		case lineKindText, lineKindListContinue:
			node.kind = nodeKindParagraph
			node.WriteString(line)
			node.WriteByte('\n')
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
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
			line, _ = doc.consumeLinesUntil(
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
			node.parseInlineMarkup(doc, nodeKindText)
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindSectionL1, nodeKindSectionL2,
			nodeKindSectionL3, nodeKindSectionL4,
			nodeKindSectionL5:
			if term == nodeKindBlockOpen {
				node.kind = nodeKindParagraph
				node.WriteString(line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
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
				node.parseInlineMarkup(doc, nodeKindText)
				parent.addChild(node)
				node = new(adocNode)
				continue
			}

			node.kind = doc.kind
			node.WriteString(
				// BUG: "= =a" could become "a", it should be "=a"
				strings.TrimLeft(line, "= \t"),
			)

			var expParent = doc.kind - 1
			for parent.kind != expParent {
				parent = parent.parent
				if parent == nil {
					parent = doc.content
					break
				}
			}
			node.parseSection(doc)
			parent.addChild(node)
			parent = node
			node = new(adocNode)
			line = ""
			continue

		case nodeKindLiteralParagraph:
			if node.IsStyleAdmonition() {
				node.kind = nodeKindParagraph
				node.WriteString(spaces + line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
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
				node.parseInlineMarkup(doc, nodeKindText)
			} else {
				node.kind = doc.kind
				node.classes = append(node.classes, classNameLiteralBlock)
				node.WriteString(line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
					node,
					lineKindEmpty,
					[]int{
						term,
						nodeKindBlockListing,
						nodeKindBlockListingNamed,
						nodeKindBlockLiteral,
						nodeKindBlockLiteralNamed,
					})
				node.applySubstitutions()
			}
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockLiteral:
			node.kind = doc.kind
			node.classes = append(node.classes, classNameLiteralBlock)
			line, _ = doc.consumeLinesUntil(node, doc.kind, nil)
			node.applySubstitutions()
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockLiteralNamed:
			node.kind = doc.kind
			node.classes = append(node.classes, classNameLiteralBlock)
			line, _ = doc.consumeLinesUntil(node, lineKindEmpty, nil)
			node.applySubstitutions()
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListing:
			node.kind = doc.kind
			node.classes = append(node.classes, classNameListingBlock)
			line, _ = doc.consumeLinesUntil(node, doc.kind, nil)
			node.applySubstitutions()
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockListingNamed:
			node.kind = doc.kind
			node.classes = append(node.classes, classNameListingBlock)
			line, _ = doc.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					nodeKindBlockListing,
					nodeKindBlockListingNamed,
					nodeKindBlockLiteral,
					nodeKindBlockLiteralNamed,
					lineKindListContinue,
				})
			node.applySubstitutions()
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockPassthrough:
			node.kind = doc.kind
			line, _ = doc.consumeLinesUntil(node, doc.kind, nil)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListOrderedItem:
			line, _ = doc.parseListOrdered(parent, node.rawTitle, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListUnorderedItem:
			line, _ = doc.parseListUnordered(parent, node, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindListDescriptionItem:
			line, _ = doc.parseListDescription(parent, node, line)
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockImage:
			lineImage := strings.TrimRight(line[7:], " \t")
			if node.parseImage(lineImage) {
				node.kind = doc.kind
				line = ""
			} else {
				node.kind = nodeKindParagraph
				node.WriteString(line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
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
				node.parseInlineMarkup(doc, nodeKindText)
			}
			parent.addChild(node)
			node = &adocNode{}
			continue

		case nodeKindBlockOpen, nodeKindBlockExample,
			nodeKindBlockSidebar:
			node.kind = doc.kind
			doc.parseBlock(node, doc.kind)
			parent.addChild(node)
			node = new(adocNode)
			line = ""
			continue

		case nodeKindBlockExcerpts:
			node.kind = doc.kind
			if node.IsStyleVerse() {
				line, _ = doc.consumeLinesUntil(
					node,
					doc.kind,
					[]int{
						term,
						nodeKindBlockListing,
						nodeKindBlockListingNamed,
						nodeKindBlockLiteral,
						nodeKindBlockLiteralNamed,
						lineKindListContinue,
					})
			} else {
				doc.parseBlock(node, doc.kind)
				line = ""
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockVideo:
			if node.parseVideo(line) {
				node.kind = doc.kind
				line = ""
			} else {
				node.kind = nodeKindParagraph
				node.WriteString("video::" + line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
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
				node.parseInlineMarkup(doc, nodeKindText)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindBlockAudio:
			if node.parseBlockAudio(line) {
				node.kind = doc.kind
				line = ""
			} else {
				node.kind = nodeKindParagraph
				node.WriteString("audio::" + line)
				node.WriteByte('\n')
				line, _ = doc.consumeLinesUntil(
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
				node.parseInlineMarkup(doc, nodeKindText)
			}
			parent.addChild(node)
			node = new(adocNode)
			continue

		case nodeKindMacroTOC:
			node.kind = doc.kind
			parent.addChild(node)
			node = new(adocNode)
		}
		line = ""
	}
}

//
// parseHeader document header consist of title and optional authors,
// revision, and zero or more attributes.
// The document attributes can be in any order, but the author and revision MUST
// be in order.
//
//	DOC_HEADER  = *(DOC_ATTRIBUTE / COMMENTS)
//	              "=" SP *ADOC_WORD LF
//	              (*DOC_ATTRIBUTE)
//	              DOC_AUTHORS LF
//	              (*DOC_ATTRIBUTE)
//	              DOC_REVISION LF
//	              (*DOC_ATTRIBUTE)
//
func (doc *Document) parseHeader() {
	const (
		stateTitle int = iota
		stateAuthor
		stateRevision
		stateEnd
	)
	state := stateTitle
	for {
		_, line, c := doc.line()
		if len(line) == 0 && c == 0 {
			break
		}
		if len(line) == 0 {
			// Only allow empty line if state is title.
			if state == stateTitle {
				continue
			}
			return
		}

		if strings.HasPrefix(line, "////") {
			doc.parseIgnoreCommentBlock()
			continue
		}
		if strings.HasPrefix(line, "//") {
			continue
		}
		if line[0] == ':' {
			key, value := doc.parseAttribute(line, false)
			if len(key) > 0 {
				doc.Attributes.apply(key, value)
				continue
			}
			if state != stateTitle {
				return
			}
			// The line will be assumed either as author or
			// revision.
		}
		switch state {
		case stateTitle:
			if !isTitle(line) {
				return
			}
			doc.header = &adocNode{
				kind: nodeKindDocHeader,
			}
			doc.header.WriteString(strings.TrimSpace(line[2:]))
			doc.Title = string(doc.header.raw)
			state = stateAuthor
		case stateAuthor:
			doc.Author = line
			state = stateRevision
		case stateRevision:
			if !doc.parseHeaderRevision(line) {
				return
			}
			state = stateEnd
		case stateEnd:
			return
		}
	}
	return
}

//
//	DOC_REVISION     = DOC_REV_VERSION [ "," DOC_REV_DATE ]
//
//	DOC_REV_VERSION  = "v" 1*DIGIT "." 1*DIGIT "." 1*DIGIT
//
//	DOC_REV_DATE     = 1*2DIGIT WSP 3*ALPHA WSP 4*DIGIT
//
func (doc *Document) parseHeaderRevision(line string) bool {
	if line[0] != 'v' {
		return false
	}

	idx := strings.IndexByte(line, ',')
	if idx > 0 {
		doc.RevNumber = line[1:idx]
		doc.RevDate = strings.TrimSpace(line[idx+1:])
		doc.RevSeparator = ","
	} else {
		doc.RevNumber = line[1:]
	}
	return true
}

func (doc *Document) parseIgnoreCommentBlock() {
	for {
		line, c := doc.p.Line()
		if strings.HasPrefix(line, "////") {
			return
		}
		if len(line) == 0 && c == 0 {
			return
		}
	}
}

//
// parseListBlock parse block after list continuation "+" until we found
// empty line or non-list line.
//
func (doc *Document) parseListBlock() (node *adocNode, line string, c rune) {
	for {
		_, line, c = doc.line()
		if len(line) == 0 && c == 0 {
			break
		}

		if doc.kind == lineKindAdmonition {
			node = &adocNode{
				kind:  nodeKindParagraph,
				style: styleAdmonition,
			}
			node.parseLineAdmonition(line)
			line, c = doc.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					nodeKindBlockListing,
					nodeKindBlockListingNamed,
					nodeKindBlockLiteral,
					nodeKindBlockLiteralNamed,
					lineKindListContinue,
				})
			node.parseInlineMarkup(doc, nodeKindText)
			break
		}
		if doc.kind == lineKindBlockComment {
			doc.parseIgnoreCommentBlock()
			continue
		}
		if doc.kind == lineKindComment {
			continue
		}
		if doc.kind == lineKindEmpty {
			return node, line, c
		}
		if doc.kind == lineKindListContinue {
			continue
		}
		if doc.kind == nodeKindLiteralParagraph {
			node = &adocNode{
				kind:    doc.kind,
				classes: []string{classNameLiteralBlock},
			}
			node.WriteString(strings.TrimLeft(line, " \t"))
			node.WriteByte('\n')
			line, c = doc.consumeLinesUntil(
				node,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			break
		}
		if doc.kind == lineKindText {
			node = &adocNode{
				kind: nodeKindParagraph,
			}
			node.WriteString(line)
			node.WriteByte('\n')
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					lineKindListContinue,
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
					nodeKindListDescriptionItem,
				})
			node.postParseParagraph(nil)
			node.parseInlineMarkup(doc, nodeKindText)
			break
		}
		if doc.kind == nodeKindBlockListing {
			node = &adocNode{
				kind:    doc.kind,
				classes: []string{classNameListingBlock},
			}
			doc.consumeLinesUntil(node, doc.kind, nil)
			node.applySubstitutions()
			line = ""
			break
		}
		if doc.kind == nodeKindListOrderedItem {
			break
		}
		if doc.kind == nodeKindListUnorderedItem {
			break
		}
		if doc.kind == nodeKindListDescriptionItem {
			break
		}
	}
	return node, line, c
}

func (doc *Document) parseListDescription(parent, node *adocNode, line string) (
	got string, c rune,
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

	line = ""
	for {
		if len(line) == 0 {
			_, line, c = doc.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}
		if doc.kind == lineKindBlockComment {
			doc.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if doc.kind == lineKindComment {
			line = ""
			continue
		}
		if doc.kind == lineKindListContinue {
			var node *adocNode
			node, line, c = doc.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if doc.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if doc.kind == nodeKindListOrderedItem {
			line, c = doc.parseListOrdered(listItem, "", line)
			continue
		}
		if doc.kind == nodeKindListUnorderedItem {
			line, c = doc.parseListUnordered(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindListDescriptionItem {
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
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListDescriptionItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}
			line, c = doc.parseListDescription(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindBlockListingNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameListingBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockLiteralNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameLiteralBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockListing ||
			doc.kind == nodeKindBlockExample ||
			doc.kind == nodeKindBlockSidebar {
			break
		}
		if doc.kind == nodeKindSectionL1 ||
			doc.kind == nodeKindSectionL2 ||
			doc.kind == nodeKindSectionL3 ||
			doc.kind == nodeKindSectionL4 ||
			doc.kind == nodeKindSectionL5 ||
			doc.kind == lineKindAttributeElement ||
			doc.kind == lineKindBlockTitle ||
			doc.kind == lineKindID ||
			doc.kind == lineKindIDShort ||
			doc.kind == lineKindText {
			if doc.prevKind == lineKindEmpty {
				break
			}
		}

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(doc, nodeKindListDescriptionItem)
	return line, c
}

//
// parseListOrdered parser the content as list until it found line that is not
// list-item.
// On success it will return non-empty line and terminator character.
//
func (doc *Document) parseListOrdered(parent *adocNode, title, line string) (
	got string, c rune,
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

	line = ""

	for {
		if len(line) == 0 {
			_, line, c = doc.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}

		if doc.kind == lineKindBlockComment {
			doc.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if doc.kind == lineKindComment {
			line = ""
			continue
		}
		if doc.kind == lineKindListContinue {
			var node *adocNode
			node, line, c = doc.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if doc.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if doc.kind == nodeKindListOrderedItem {
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
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListOrderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListOrdered(listItem, "", line)
			continue
		}
		if doc.kind == nodeKindListUnorderedItem {
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
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListOrderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListUnordered(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindListDescriptionItem {
			node := &adocNode{
				kind: doc.kind,
			}
			node.parseListDescriptionItem(line)

			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListOrderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListDescription(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindBlockListingNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameListingBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockLiteralNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameLiteralBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockListing ||
			doc.kind == nodeKindBlockExample ||
			doc.kind == nodeKindBlockSidebar {
			break
		}
		if doc.kind == nodeKindSectionL1 ||
			doc.kind == nodeKindSectionL2 ||
			doc.kind == nodeKindSectionL3 ||
			doc.kind == nodeKindSectionL4 ||
			doc.kind == nodeKindSectionL5 ||
			doc.kind == lineKindAdmonition ||
			doc.kind == lineKindAttributeElement ||
			doc.kind == lineKindBlockTitle ||
			doc.kind == lineKindID ||
			doc.kind == lineKindIDShort ||
			doc.kind == lineKindText {
			if doc.prevKind == lineKindEmpty {
				break
			}
		}

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(doc, nodeKindListOrderedItem)
	return line, c
}

func (doc *Document) parseListUnordered(parent, node *adocNode, line string) (
	got string, c rune,
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

	line = ""

	for {
		if len(line) == 0 {
			_, line, c = doc.line()
			if len(line) == 0 && c == 0 {
				break
			}
		}

		if doc.kind == lineKindBlockComment {
			doc.parseIgnoreCommentBlock()
			line = ""
			continue
		}
		if doc.kind == lineKindComment {
			line = ""
			continue
		}
		if doc.kind == lineKindListContinue {
			var node *adocNode
			node, line, c = doc.parseListBlock()
			if node != nil {
				listItem.addChild(node)
			}
			continue
		}
		if doc.kind == lineKindEmpty {
			// Keep going, maybe next line is still a list.
			continue
		}
		if doc.kind == nodeKindListOrderedItem {
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
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListUnorderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListOrdered(listItem, "", line)
			continue
		}

		if doc.kind == nodeKindListUnorderedItem {
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
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListUnorderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListUnordered(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindListDescriptionItem {
			node := &adocNode{
				kind: doc.kind,
			}
			node.parseListDescriptionItem(line)

			parentListItem := parent
			for parentListItem != nil {
				if parentListItem.kind == doc.kind && parentListItem.level == node.level {
					list.postParseList(doc, nodeKindListUnorderedItem)
					return line, c
				}
				parentListItem = parentListItem.parent
			}

			line, c = doc.parseListDescription(listItem, node, line)
			continue
		}
		if doc.kind == nodeKindBlockListingNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameListingBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockLiteralNamed {
			if doc.prevKind == lineKindEmpty {
				break
			}
			node := &adocNode{
				kind:    doc.kind,
				classes: []string{classNameLiteralBlock},
			}
			line, c = doc.consumeLinesUntil(node,
				lineKindEmpty,
				[]int{
					nodeKindListOrderedItem,
					nodeKindListUnorderedItem,
				})
			node.applySubstitutions()
			listItem.addChild(node)
			continue
		}
		if doc.kind == nodeKindBlockListing ||
			doc.kind == nodeKindBlockExample ||
			doc.kind == nodeKindBlockSidebar {
			break
		}
		if doc.kind == nodeKindSectionL1 ||
			doc.kind == nodeKindSectionL2 ||
			doc.kind == nodeKindSectionL3 ||
			doc.kind == nodeKindSectionL4 ||
			doc.kind == nodeKindSectionL5 ||
			doc.kind == lineKindAdmonition ||
			doc.kind == lineKindAttributeElement ||
			doc.kind == lineKindBlockTitle ||
			doc.kind == lineKindID ||
			doc.kind == lineKindIDShort ||
			doc.kind == lineKindText {
			if doc.prevKind == lineKindEmpty {
				break
			}
		}

		listItem.WriteString(strings.TrimSpace(line))
		listItem.WriteByte('\n')
		line = ""
	}
	list.postParseList(doc, nodeKindListUnorderedItem)
	return line, c
}

//
// registerAnchor register ID and its label.
// If the ID is already exist it will generate new ID with additional suffix
// "_x" added, where x is the counter of duplicate ID.
// The old or new ID will be returned to caller.
//
func (doc *Document) registerAnchor(id, label string) string {
	got, ok := doc.anchors[id]
	for ok {
		log.Printf("registerAnchor duplicate: %s %+v", id, got)
		// The ID is duplicate
		got.counter++
		id = fmt.Sprintf("%s_%d", id, got.counter)
		got, ok = doc.anchors[id]
	}
	doc.anchors[id] = &anchor{
		label: label,
	}
	return id
}

//
// tocHTML write table of contents with HTML template into out.
//
func (doc *Document) tocHTML(tmpl *template.Template, out io.Writer) (err error) {
	v, ok := doc.Attributes[metaNameTOCLevels]
	if ok {
		doc.TOCLevel, err = strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("tocHTML: %w", err)
		}
		if doc.TOCLevel <= 0 {
			doc.TOCLevel = defTOCLevel
		}
	}

	v, ok = doc.Attributes[metaNameTOCTitle]
	if ok && len(v) > 0 {
		doc.TOCTitle = v
	}

	err = tmpl.ExecuteTemplate(out, "BEGIN_TOC", doc)
	if err != nil {
		return fmt.Errorf("tocHTML: BEGIN_TOC: %w", err)
	}

	err = doc.htmlGenerateTOC(doc.content, tmpl, out, 0)
	if err != nil {
		return fmt.Errorf("tocHTML: %w", err)
	}

	err = tmpl.ExecuteTemplate(out, "END_TOC", doc)
	if err != nil {
		return fmt.Errorf("tocHTML: END_TOC: %w", err)
	}

	return nil
}
