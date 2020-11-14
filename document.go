// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
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

	doc, err = Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("Open %s: %w", file, err)
	}

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

	_, err = fmt.Fprintf(out, _htmlBegin)
	if err != nil {
		return err
	}

	metaValue := doc.Attributes[metaNameDescription]
	if len(metaValue) > 0 {
		_, err = fmt.Fprintf(out, _htmlMetaDescription, metaValue)
		if err != nil {
			return err
		}
	}

	metaValue = doc.Attributes[metaNameKeywords]
	if len(metaValue) > 0 {
		_, err = fmt.Fprintf(out, _htmlMetaKeywords, metaValue)
		if err != nil {
			return err
		}
	}

	if len(doc.Author) > 0 {
		_, err = fmt.Fprintf(out, _htmlMetaAuthor, doc.Author)
		if err != nil {
			return err
		}
	}

	title := doc.Attributes[metaNameTitle]
	if len(title) == 0 && len(doc.Title) > 0 {
		title = doc.Title
	}
	if len(title) > 0 {
		_, err = fmt.Fprintf(out, _htmlHeadTitle, title)
		if err != nil {
			return err
		}
	}
	if _, err = fmt.Fprint(out, _htmlHeadStyle); err != nil {
		return err
	}

	bodyClasses := strings.Join(doc.classes, " ")
	if _, err = fmt.Fprintf(out, _htmlBodyBegin, bodyClasses); err != nil {
		return err
	}

	if err = htmlWriteBody(doc, out); err != nil {
		return err
	}

	if _, err = fmt.Fprint(out, _htmlFooterBegin); err != nil {
		return err
	}
	if len(doc.RevNumber) > 0 {
		_, err = fmt.Fprintf(out, _htmlFooterVersion, doc.RevNumber)
		if err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(out, _htmlFooterLastUpdated, doc.LastUpdated); err != nil {
		if err != nil {
			return err
		}
	}
	if _, err = fmt.Fprint(out, _htmlFooterEnd); err != nil {
		return err
	}
	_, err = fmt.Fprint(out, _htmlBodyEnd)

	return err
}

//
// ToHTMLBody convert the document object into HTML with content of body only.
//
func (doc *Document) ToHTMLBody(w io.Writer) (err error) {
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

	err = htmlWriteBody(doc, w)
	if err != nil {
		return err
	}

	return nil
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
func (doc *Document) tocHTML(out io.Writer) (err error) {
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

	tocClasses := strings.Join(doc.tocClasses, " ")
	v, ok = doc.Attributes[metaNameTOCTitle]
	if ok && len(v) > 0 {
		doc.tocTitle = v
	}

	_, err = fmt.Fprintf(out, _htmlToCBegin, tocClasses, doc.tocTitle)
	if err != nil {
		return fmt.Errorf("tocHTML: _htmlToCBegin: %w", err)
	}

	err = htmlWriteToC(doc, doc.content, out, 0)
	if err != nil {
		return fmt.Errorf("tocHTML: %w", err)
	}

	_, err = fmt.Fprintf(out, _htmlToCEnd)
	if err != nil {
		return fmt.Errorf("tocHTML: _htmlToCEnd: %w", err)
	}

	return nil
}
