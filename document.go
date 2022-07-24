// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// Document represent content of asciidoc that has been parsed.
type Document struct {
	// anchors contains mapping between unique ID and its label.
	anchors map[string]*anchor

	header   *element
	preamble *element
	content  *element

	Attributes AttributeEntry
	sectnums   *sectionCounters

	// titleID is the reverse of anchors, it contains mapping of title and
	// its ID.
	titleID map[string]string

	Revision Revision

	LastUpdated string
	file        string
	fpath       string
	rawAuthors  string
	rawRevision string
	tocPosition string
	tocTitle    string

	Title DocumentTitle

	classes    attributeClass
	tocClasses attributeClass

	Authors []*Author

	TOCLevel       int
	sectLevel      int
	counterExample int
	counterImage   int
	counterTable   int

	isEmbedded   bool
	isForToC     bool
	tocIsEnabled bool
}

func newDocument() (doc *Document) {
	doc = &Document{
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
		header: &element{
			kind: elKindDocHeader,
		},
		preamble: &element{
			elementAttribute: elementAttribute{
				Attrs: make(map[string]string),
			},
			kind: elKindPreamble,
		},
		content: &element{
			kind: elKindDocContent,
		},
	}

	doc.content.addChild(doc.preamble)

	return doc
}

// Open the ascidoc file and parse it.
func Open(file string) (doc *Document, err error) {
	var (
		fi  os.FileInfo
		wd  string
		raw []byte
	)

	fi, err = os.Stat(file)
	if err != nil {
		return nil, err
	}

	raw, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Open %s: %w", file, err)
	}

	wd, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Open %s: %w", file, err)
	}

	doc = newDocument()
	doc.fpath = filepath.Join(wd, file)
	doc.file = file
	doc.LastUpdated = fi.ModTime().Round(time.Second).Format("2006-01-02 15:04:05 Z0700")

	parse(doc, raw)

	return doc, nil
}

// Parse the content into a Document.
func Parse(content []byte) (doc *Document) {
	doc = newDocument()
	parse(doc, content)
	return doc
}

func parse(doc *Document, content []byte) {
	var (
		docp *documentParser = newDocumentParser(doc, content)

		sectLevel string
		ok        bool
	)

	docp.parseHeader()
	docp.doc.postParseHeader()

	sectLevel, ok = doc.Attributes[metaNameSectNumLevel]
	if ok {
		doc.sectLevel, _ = strconv.Atoi(sectLevel)
	}

	docp.parseBlock(doc.preamble, 0)
}

// ToHTMLEmbedded convert the Document object into HTML with content only,
// without header and footer.
func (doc *Document) ToHTMLEmbedded(out io.Writer) (err error) {
	doc.isEmbedded = true
	doc.generateClasses()

	var buf = &bytes.Buffer{}
	doc.toHTMLBody(buf, false)
	doc.isEmbedded = false
	_, err = out.Write(buf.Bytes())
	return err
}

// ToHTML convert the Document object into full HTML document.
func (doc *Document) ToHTML(out io.Writer) (err error) {
	var (
		metaValue string
	)

	doc.generateClasses()

	// Use *bytes.Buffer to minimize checking for error.
	var buf = &bytes.Buffer{}

	metaValue = doc.Attributes[MetaNameGenerator]

	fmt.Fprintf(buf, _htmlBegin, metaValue)

	metaValue = doc.Attributes[MetaNameDescription]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=\"description\" content=%q>",
			metaValue)
	}

	metaValue = doc.Attributes[MetaNameKeywords]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=\"keywords\" content=%q>", metaValue)
	}

	metaValue = doc.Attributes[MetaNameAuthorNames]
	if len(metaValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>", MetaNameAuthor, metaValue)
	}

	var title string = doc.Title.String()
	if len(title) > 0 {
		fmt.Fprintf(buf, "\n<title>%s</title>", title)
	}
	fmt.Fprint(buf, "\n<style>\n\n</style>")

	fmt.Fprintf(buf, "\n</head>\n<body class=%q>", doc.classes.String())

	var (
		isWithHeaderFooter = true
		ok                 bool
	)
	_, ok = doc.Attributes[metaNameNoHeaderFooter]
	if ok {
		isWithHeaderFooter = false
	}
	doc.toHTMLBody(buf, isWithHeaderFooter)

	fmt.Fprint(buf, "\n</body>\n</html>")

	_, err = out.Write(buf.Bytes())

	return err
}

// ToHTMLBody convert the Document object into HTML with body only, this is
// including header, content, and footer.
func (doc *Document) ToHTMLBody(out io.Writer) (err error) {
	doc.generateClasses()

	var buf = &bytes.Buffer{}
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
	var (
		ok bool
	)

	if withHeaderFooter {
		_, ok = doc.Attributes[metaNameNoHeader]
		if !ok {
			htmlWriteHeader(doc, buf)
		}
	}

	htmlWriteBody(doc, buf)

	if withHeaderFooter {
		_, ok = doc.Attributes[metaNameNoFooter]
		if !ok {
			htmlWriteFooter(doc, buf)
		}
	}
}

// postParseHeader re-check the document title, substract the authors, and
// revision number, date, and/or remark.
func (doc *Document) postParseHeader() {
	doc.unpackTitleSeparator()
	doc.unpackRawTitle()
	doc.unpackRawAuthor()
	doc.unpackRawRevision()
}

// registerAnchor register ID and its label.
// If the ID is already exist it will generate new ID with additional suffix
// "_x" added, where x is the counter of duplicate ID.
// The old or new ID will be returned to caller.
func (doc *Document) registerAnchor(id, label string) string {
	var (
		got *anchor
		ok  bool
	)

	got, ok = doc.anchors[id]
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

// tocHTML write table of contents with HTML template into out.
func (doc *Document) tocHTML(out io.Writer) {
	var (
		v  string
		ok bool
	)

	v, ok = doc.Attributes[metaNameTOCLevels]
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

// unpackRawAuthor parse the authors field into one or more Author.
//
// This method set the Document Attributes "author_names" to list of author
// full name, separated by comma.
//
// If the Authors is more than one, the first author is set using Attributes
// "author", "author_1, "email", "email_1", and so on;
// and the second author is set using "author_2, "email_2", and so on.
func (doc *Document) unpackRawAuthor() {
	var (
		sb strings.Builder
		v  string
	)

	if len(doc.rawAuthors) == 0 {
		v = doc.Attributes[MetaNameAuthor]
		if len(v) > 0 {
			sb.WriteString(v)
		}
		v = doc.Attributes[metaNameEmail]
		if len(v) > 0 {
			sb.WriteString(" <")
			sb.WriteString(v)
			sb.WriteString(">")
		}
		v = sb.String()
		if len(v) == 0 {
			return
		}
		doc.rawAuthors = v
	}

	var (
		rawAuthors    []string = strings.Split(doc.rawAuthors, ";")
		authorKey              = MetaNameAuthor
		emailKey               = metaNameEmail
		initialsKey            = metaNameAuthorInitials
		firstNameKey           = metaNameFirstName
		middleNameKey          = metaNameMiddleName
		lastNameKey            = metaNameLastName

		author    *Author
		rawAuthor string
		x         int
	)

	sb.Reset()

	for x, rawAuthor = range rawAuthors {
		if len(rawAuthor) == 0 {
			continue
		}

		author = parseAuthor(rawAuthor)
		doc.Authors = append(doc.Authors, author)

		if len(doc.Authors) >= 2 {
			sb.WriteString(", ")
		}
		sb.WriteString(author.FullName())

		if x == 0 {
			doc.Attributes[authorKey] = author.FullName()
			doc.Attributes[emailKey] = author.Email
			doc.Attributes[initialsKey] = author.Initials
			doc.Attributes[firstNameKey] = author.FirstName
			doc.Attributes[middleNameKey] = author.MiddleName
			doc.Attributes[lastNameKey] = author.LastName

			// No continue, the first author have two keys, one is
			// "author" and another is "author_1".
		}

		authorKey = fmt.Sprintf("%s_%d", MetaNameAuthor, x+1)
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

	v = sb.String()
	if len(v) > 0 {
		doc.Attributes[MetaNameAuthorNames] = v
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
		x     int
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

	doc.Title.el = parseInlineMarkup(doc, []byte(doc.Title.raw))
	title = doc.Title.el.toText()
	doc.Attributes[metaNameDocTitle] = title

	for x = len(title) - 1; x > 0; x-- {
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

// unpackTitleSeparator set the Title separator using the first character in
// meta attribute "title-separator" value.
func (doc *Document) unpackTitleSeparator() {
	var (
		v  string
		ok bool
	)

	v, ok = doc.Attributes[metaNameTitleSeparator]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			doc.Title.sep = v[0]
		}
	}
}
