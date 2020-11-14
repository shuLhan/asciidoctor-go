// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

//
// Document represent content of asciidoc that has been parsed.
//
type Document struct {
	Author       string
	Title        string
	RevNumber    string
	RevSeparator string
	RevDate      string
	LastUpdated  string
	Attributes   AttributeEntry

	TOCLevel     int
	tocClasses   []string
	tocPosition  string
	tocTitle     string
	tocIsEnabled bool

	file    string
	classes []string

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

	counterImage   int
	counterExample int
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

	doc = Parse(raw)
	doc.file = file
	doc.LastUpdated = fi.ModTime().Round(time.Second).Format("2006-01-02 15:04:05 Z0700")

	return doc, nil
}

//
// ToHTML convert the asciidoc document into full HTML document, including
// head and body.
//
func (doc *Document) ToHTML(out io.Writer) (err error) {
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

	// Use *bytes.Buffer to minimize checking for error.
	buf := &bytes.Buffer{}

	fmt.Fprint(buf, _htmlBegin)

	metaValue := doc.Attributes[metaNameDescription]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, _htmlMetaDescription, metaValue)
	}

	metaValue = doc.Attributes[metaNameKeywords]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, _htmlMetaKeywords, metaValue)
	}

	if len(doc.Author) > 0 {
		fmt.Fprintf(buf, _htmlMetaAuthor, doc.Author)
	}

	title := doc.Attributes[metaNameTitle]
	if len(title) == 0 && len(doc.Title) > 0 {
		title = doc.Title
	}
	if len(title) > 0 {
		fmt.Fprintf(buf, _htmlHeadTitle, title)
	}
	fmt.Fprint(buf, _htmlHeadStyle)

	bodyClasses := strings.Join(doc.classes, " ")
	fmt.Fprintf(buf, _htmlBodyBegin, bodyClasses)

	htmlWriteBody(doc, buf)

	fmt.Fprint(buf, _htmlFooterBegin)
	if len(doc.RevNumber) > 0 {
		fmt.Fprintf(buf, _htmlFooterVersion, doc.RevNumber)
	}
	fmt.Fprintf(buf, _htmlFooterLastUpdated, doc.LastUpdated)
	fmt.Fprint(buf, _htmlFooterEnd)
	fmt.Fprint(buf, _htmlBodyEnd)

	_, err = out.Write(buf.Bytes())

	return err
}

//
// ToHTMLBody convert the document object into HTML with content of body only.
//
func (doc *Document) ToHTMLBody(out io.Writer) (err error) {
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

	buf := &bytes.Buffer{}
	htmlWriteBody(doc, buf)
	_, err = out.Write(buf.Bytes())

	return err
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
func (doc *Document) tocHTML(out io.Writer) {
	v, ok := doc.Attributes[metaNameTOCLevels]
	if ok {
		doc.TOCLevel, _ = strconv.Atoi(v)
		if doc.TOCLevel <= 0 {
			doc.TOCLevel = defTOCLevel
		}
	}

	tocClasses := strings.Join(doc.tocClasses, " ")
	v, ok = doc.Attributes[metaNameTOCTitle]
	if ok && len(v) > 0 {
		doc.tocTitle = v
	}

	fmt.Fprintf(out, _htmlToCBegin, tocClasses, doc.tocTitle)
	htmlWriteToC(doc, doc.content, out, 0)
	fmt.Fprint(out, _htmlToCEnd)
}
