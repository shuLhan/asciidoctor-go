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

const (
	defSectnumlevels  = 3
	defTOCLevel       = 2
	defTOCTitle       = "Table of Contents"
	defTitleSeparator = ':'
	defVersionPrefix  = "version "
)

//
// Document represent content of asciidoc that has been parsed.
//
type Document struct {
	Title       DocumentTitle
	Authors     []*Author
	rawAuthors  string
	Revision    Revision
	rawRevision string
	LastUpdated string
	Attributes  AttributeEntry

	TOCLevel     int
	tocClasses   attributeClass
	tocPosition  string
	tocTitle     string
	tocIsEnabled bool

	file    string
	classes attributeClass

	// anchors contains mapping between unique ID and its label.
	anchors map[string]*anchor
	// titleID is the reverse of anchors, it contains mapping of title and
	// its ID.
	titleID map[string]string

	sectnums  *sectionCounters
	sectLevel int

	header  *adocNode
	content *adocNode

	counterExample int
	counterImage   int
	counterTable   int
}

func newDocument() *Document {
	return &Document{
		Title: DocumentTitle{
			sep: defTitleSeparator,
		},
		TOCLevel:   defTOCLevel,
		tocClasses: attributeClass{},
		tocTitle:   defTOCTitle,
		Attributes: newAttributeEntry(),
		classes:    attributeClass{},
		anchors:    make(map[string]*anchor),
		titleID:    make(map[string]string),
		sectnums:   &sectionCounters{},
		sectLevel:  defSectnumlevels,
		header: &adocNode{
			kind: nodeKindDocHeader,
		},
		content: &adocNode{
			kind: nodeKindDocContent,
		},
	}
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
// ToEmbeddedHTML convert the Document object into HTML with content only
// (without header and footer).
//
func (doc *Document) ToEmbeddedHTML(out io.Writer) (err error) {
	doc.generateClasses()
	buf := &bytes.Buffer{}
	doc.toHTMLBody(buf, false)
	_, err = out.Write(buf.Bytes())
	return err
}

//
// ToHTML convert the asciidoc document into full HTML document, including
// head and body.
//
func (doc *Document) ToHTML(out io.Writer) (err error) {
	doc.generateClasses()

	// Use *bytes.Buffer to minimize checking for error.
	buf := &bytes.Buffer{}

	fmt.Fprint(buf, _htmlBegin)

	metaValue := doc.Attributes[metaNameDescription]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=\"description\" content=%q>",
			metaValue)
	}

	metaValue = doc.Attributes[metaNameKeywords]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=\"keywords\" content=%q>", metaValue)
	}

	var metaAuthors strings.Builder
	for x, author := range doc.Authors {
		if x > 0 {
			metaAuthors.WriteString(", ")
		}
		metaAuthors.WriteString(author.FullName())
	}
	if metaAuthors.Len() > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>",
			attrValueAuthor, metaAuthors.String())
	}

	title := doc.Title.String()
	if len(title) > 0 {
		fmt.Fprintf(buf, "\n<title>%s</title>", title)
	}
	fmt.Fprint(buf, "\n<style>\n\n</style>")

	fmt.Fprintf(buf, "\n</head>\n<body class=%q>", doc.classes.String())

	isWithHeaderFooter := true
	_, ok := doc.Attributes[metaNameNoHeaderFooter]
	if ok {
		isWithHeaderFooter = false
	}
	doc.toHTMLBody(buf, isWithHeaderFooter)

	fmt.Fprint(buf, "\n</body>\n</html>")

	_, err = out.Write(buf.Bytes())

	return err
}

//
// ToHTMLBody convert the Document object body into HTML, this is including
// header, content, and footer.
//
func (doc *Document) ToHTMLBody(out io.Writer) (err error) {
	doc.generateClasses()
	buf := &bytes.Buffer{}
	doc.toHTMLBody(buf, true)
	_, err = out.Write(buf.Bytes())
	return err
}

func (doc *Document) generateClasses() {
	doc.classes.add(classNameArticle)
	doc.tocPosition, doc.tocIsEnabled = doc.Attributes[metaNameTOC]

	switch doc.tocPosition {
	case metaValueLeft:
		doc.classes.add(classNameToc2)
		doc.classes.add(classNameTocLeft)
		doc.tocClasses.add(classNameToc2)
	case metaValueRight:
		doc.classes.add(classNameToc2)
		doc.classes.add(classNameTocRight)
		doc.tocClasses.add(classNameToc2)
	default:
		doc.tocClasses.add(classNameToc)
	}
}

func (doc *Document) toHTMLBody(buf *bytes.Buffer, withHeaderFooter bool) {
	if withHeaderFooter {
		_, ok := doc.Attributes[metaNameNoHeader]
		if !ok {
			htmlWriteHeader(doc, buf)
		}
	}

	htmlWriteBody(doc, buf)

	if withHeaderFooter {
		_, ok := doc.Attributes[metaNameNoFooter]
		if !ok {
			htmlWriteFooter(doc, buf)
		}
	}
}

//
// postParseHeader re-check the document title, substract the authors, and
// revision number, date, and/or remark.
//
func (doc *Document) postParseHeader() {
	doc.unpackTitleSeparator()
	doc.unpackRawTitle()
	doc.unpackRawAuthor()
	doc.unpackRawRevision()
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

	v, ok = doc.Attributes[metaNameTOCTitle]
	if ok && len(v) > 0 {
		doc.tocTitle = v
	}

	fmt.Fprintf(out, _htmlToCBegin, doc.tocClasses.String(), doc.tocTitle)
	htmlWriteToC(doc, doc.content, out, 0)
	fmt.Fprint(out, "\n</div>")
}

//
// unpackRawAuthor parse the authors field into one or more Author.
//
func (doc *Document) unpackRawAuthor() {
	if len(doc.rawAuthors) == 0 {
		v := doc.Attributes[metaNameAuthor]
		if len(v) > 0 {
			doc.rawAuthors = v
		}
		v = doc.Attributes[metaNameEmail]
		if len(v) > 0 {
			doc.rawAuthors += " <" + v + ">"
		}
		if len(doc.rawAuthors) == 0 {
			return
		}
	}

	rawAuthors := strings.Split(doc.rawAuthors, ";")
	for _, rawAuthor := range rawAuthors {
		if len(rawAuthor) > 0 {
			doc.Authors = append(doc.Authors, parseAuthor(rawAuthor))
		}
	}

	authorKey := metaNameAuthor
	emailKey := metaNameEmail
	initialsKey := metaNameAuthorInitials
	firstNameKey := metaNameFirstName
	middleNameKey := metaNameMiddleName
	lastNameKey := metaNameLastName
	for x, author := range doc.Authors {
		if x == 0 {
			doc.Attributes[authorKey] = author.FullName()
			doc.Attributes[emailKey] = author.Email
			doc.Attributes[initialsKey] = author.Initials
			doc.Attributes[firstNameKey] = author.FirstName
			doc.Attributes[middleNameKey] = author.MiddleName
			doc.Attributes[lastNameKey] = author.LastName
		}

		authorKey = fmt.Sprintf("%s_%d", metaNameAuthor, x+1)
		emailKey = fmt.Sprintf("%s_%d", metaNameEmail, x+1)
		initialsKey = fmt.Sprintf("%s_%d", metaNameAuthorInitials, x+1)
		firstNameKey = fmt.Sprintf("%s_%d", metaNameFirstName, x+1)
		middleNameKey = fmt.Sprintf("%s_%d", metaNameMiddleName, x+1)
		lastNameKey = fmt.Sprintf("%s_%d", metaNameLastName, x+1)

		doc.Attributes[authorKey] = author.FullName()
		doc.Attributes[emailKey] = author.Email
		doc.Attributes[initialsKey] = author.Initials
		doc.Attributes[firstNameKey] = author.FirstName
		doc.Attributes[middleNameKey] = author.MiddleName
		doc.Attributes[lastNameKey] = author.LastName
	}
}

func (doc *Document) unpackRawRevision() {
	if len(doc.rawRevision) > 0 {
		doc.Revision = parseRevision(doc.rawRevision)
		doc.Attributes[metaNameRevNumber] = doc.Revision.Number
		doc.Attributes[metaNameRevDate] = doc.Revision.Date
		doc.Attributes[metaNameRevRemark] = doc.Revision.Remark
		return
	}
	doc.Revision.Number = doc.Attributes[metaNameRevNumber]
	doc.Revision.Date = doc.Attributes[metaNameRevDate]
	doc.Revision.Remark = doc.Attributes[metaNameRevRemark]
}

func (doc *Document) unpackRawTitle() {
	var (
		title string
		prev  byte
	)

	if len(doc.Title.raw) == 0 {
		doc.Title.raw = doc.Attributes[metaNameDocTitle]
		if len(doc.Title.raw) == 0 {
			doc.Title.raw = doc.Attributes[metaNameTitle]
			if len(doc.Title.raw) == 0 {
				return
			}
		}
	}

	doc.Title.node = parseInlineMarkup(doc, []byte(doc.Title.raw))
	title = doc.Title.node.toText()
	doc.Attributes[metaNameDocTitle] = title

	for x := len(title) - 1; x > 0; x-- {
		if title[x] == doc.Title.sep {
			if prev == ' ' {
				doc.Title.Sub = string(title[x+2:])
				doc.Title.Main = string(title[:x])
				break
			}
		}
		prev = title[x]
	}
	if len(doc.Title.Main) == 0 {
		doc.Title.Main = title
	}
}

//
// unpackTitleSeparator set the Title separator using the first character in
// meta attribute "title-separator" value.
//
func (doc *Document) unpackTitleSeparator() {
	v, ok := doc.Attributes[metaNameTitleSeparator]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			doc.Title.sep = v[0]
		}
	}
}
