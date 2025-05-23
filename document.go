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
	defTOCTitle       = `Table of Contents`
	defTitleSeparator = ':'
	defVersionPrefix  = `version `
)

// Document represent content of asciidoc that has been parsed.
type Document struct {
	// anchors contains mapping between unique ID and its label.
	anchors map[string]*anchor

	header   *element
	preamble *element
	content  *element

	Attributes DocumentAttribute
	sectnums   *sectionCounters

	// titleID is the reverse of anchors, it contains mapping of title and
	// its ID.
	titleID map[string]string

	// List of footnote ID and its text.
	footnotes []*macro

	Revision Revision

	file string

	// docdir contains the directory where file located.
	// This is the default value when ":docdir:" attribute is set and
	// empty.
	docdir string

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
		Attributes: newDocumentAttribute(),
		classes:    attributeClass{},
		anchors:    make(map[string]*anchor),
		titleID:    make(map[string]string),
		sectnums:   &sectionCounters{},
		sectLevel:  defSectnumlevels,
		header: &element{
			kind: elKindDocHeader,
		},
		content: &element{
			kind: elKindDocContent,
		},
	}
	return doc
}

// Open the ascidoc file and parse it.
func Open(file string) (doc *Document, err error) {
	var (
		fi  os.FileInfo
		raw []byte
	)

	fi, err = os.Stat(file)
	if err != nil {
		return nil, err
	}

	raw, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`Open %s: %w`, file, err)
	}

	doc = newDocument()
	doc.file = file
	doc.docdir = filepath.Dir(file)

	var modTime = fi.ModTime().Round(time.Second).Format(`2006-01-02 15:04:05 Z0700`)
	doc.Attributes.Entry[docAttrLastUpdateValue] = modTime
	doc.Attributes.Entry[docAttrDocdir] = doc.docdir

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
		docp = newDocumentParser(doc, content)

		sectLevel string
		ok        bool
	)

	docp.parseHeader()
	docp.doc.postParseHeader()

	sectLevel, ok = doc.Attributes.Entry[docAttrSectNumLevel]
	if ok {
		doc.sectLevel, _ = strconv.Atoi(sectLevel)
	}

	if docp.hasPreamble() {
		doc.preamble = &element{
			elementAttribute: elementAttribute{
				Attrs: make(map[string]string),
			},
			kind: elKindPreamble,
		}
		docp.parseBlock(doc.preamble, 0)
	}
	docp.parseBlock(doc.content, 0)
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
		docAttrValue string
	)

	doc.generateClasses()

	// Use *bytes.Buffer to minimize checking for error.
	var buf = &bytes.Buffer{}

	fmt.Fprint(buf, _htmlBegin)

	docAttrValue = doc.Attributes.Entry[DocAttrGenerator]
	if len(docAttrValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>", DocAttrGenerator, docAttrValue)
	}

	docAttrValue = doc.Attributes.Entry[DocAttrDescription]
	if len(docAttrValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>", DocAttrDescription, docAttrValue)
	}

	docAttrValue = doc.Attributes.Entry[DocAttrKeywords]
	if len(docAttrValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>", DocAttrKeywords, docAttrValue)
	}

	docAttrValue = doc.Attributes.Entry[DocAttrAuthorNames]
	if len(docAttrValue) > 0 {
		fmt.Fprintf(buf, "\n<meta name=%q content=%q>", DocAttrAuthor, docAttrValue)
	}

	var title = doc.Title.String()
	if len(title) > 0 {
		fmt.Fprintf(buf, "\n<title>%s</title>", title)
	}

	var ok bool

	docAttrValue, ok = doc.Attributes.Entry[DocAttrStylesheet]
	if ok && len(docAttrValue) == 0 {
		buf.WriteByte('\n')
		buf.WriteString(_defaultCSS)
	}

	fmt.Fprintf(buf, "\n</head>\n<body class=%q>\n", doc.classes.String())

	var isWithHeaderFooter = true

	_, ok = doc.Attributes.Entry[docAttrNoHeaderFooter]
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
	doc.tocPosition, doc.tocIsEnabled = doc.Attributes.Entry[docAttrTOC]

	switch doc.tocPosition {
	case docAttrValueLeft:
		doc.classes.add(classNameToc2)
		doc.classes.add(classNameTocLeft)
		doc.tocClasses.add(classNameToc2)
	case docAttrValueRight:
		doc.classes.add(classNameToc2)
		doc.classes.add(classNameTocRight)
		doc.tocClasses.add(classNameToc2)
	default:
		doc.tocClasses.add(classNameToc)
	}
}

func (doc *Document) haveHeader() bool {
	if len(doc.Authors) > 0 {
		return true
	}
	if len(doc.Revision.Number) > 0 {
		return true
	}
	if len(doc.Revision.Date) > 0 {
		return true
	}
	if len(doc.Revision.Remark) > 0 {
		return true
	}
	return false
}

// setAttribute store the document attribute val by its key.
func (doc *Document) setAttribute(key, val string) (err error) {
	if key[0] == '!' {
		key = strings.TrimSpace(key[1:])
		delete(doc.Attributes.Entry, key)
		return nil
	}
	var n = len(key)
	if key[n-1] == '!' {
		key = strings.TrimSpace(key[:n-1])
		delete(doc.Attributes.Entry, key)
		return nil
	}

	val = strings.TrimSpace(val)

	switch key {
	case docAttrLevelOffset:
		var offset int64
		offset, err = strconv.ParseInt(val, 10, 32)
		if err != nil {
			return fmt.Errorf(`Document: setAttribute: %s invalid value %q`, key, val)
		}
		if val[0] == '+' || val[0] == '-' {
			doc.Attributes.LevelOffset += int(offset)
		} else {
			doc.Attributes.LevelOffset = int(offset)
		}
		doc.Attributes.Entry[key] = val

	case docAttrDocdir:
		if val == `` {
			doc.Attributes.Entry[key] = doc.docdir
		} else {
			doc.Attributes.Entry[key] = val
		}

	default:
		doc.Attributes.Entry[key] = val
	}

	return nil
}

func (doc *Document) toHTMLBody(buf *bytes.Buffer, withHeaderFooter bool) {
	var (
		ok bool
	)

	if withHeaderFooter {
		_, ok = doc.Attributes.Entry[docAttrNoHeader]
		if !ok {
			htmlWriteHeader(doc, buf)
		}
	}

	htmlWriteBody(doc, buf)

	htmlWriteFootnoteDefs(doc, buf)

	if withHeaderFooter {
		_, ok = doc.Attributes.Entry[docAttrNoFooter]
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
		id = fmt.Sprintf(`%s_%d`, id, got.counter)
		got, ok = doc.anchors[id]
	}
	doc.anchors[id] = &anchor{
		label: label,
	}
	return id
}

// registerFootnote add footnote with id and text, where id is optional.
// If the id already exist it will return true.
func (doc *Document) registerFootnote(id string, rawContent []byte) (mcr *macro, exist bool) {
	if len(id) != 0 {
		// Find existing footnote with the same ID.
		for _, mcr = range doc.footnotes {
			if mcr.key == id {
				return mcr, true
			}
		}
	}

	mcr = &macro{
		key:        id,
		rawContent: rawContent,
	}
	doc.footnotes = append(doc.footnotes, mcr)

	mcr.level = len(doc.footnotes)

	return mcr, false
}

// tocHTML write table of contents with HTML template into out.
func (doc *Document) tocHTML(out io.Writer) {
	var (
		v  string
		ok bool
	)

	v, ok = doc.Attributes.Entry[docAttrTOCLevels]
	if ok {
		doc.TOCLevel, _ = strconv.Atoi(v)
		if doc.TOCLevel <= 0 {
			doc.TOCLevel = defTOCLevel
		}
	}

	v, ok = doc.Attributes.Entry[docAttrTOCTitle]
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
		v = doc.Attributes.Entry[DocAttrAuthor]
		if len(v) > 0 {
			sb.WriteString(v)
		}
		v = doc.Attributes.Entry[docAttrEmail]
		if len(v) > 0 {
			sb.WriteString(` <`)
			sb.WriteString(v)
			sb.WriteString(`>`)
		}
		v = sb.String()
		if len(v) == 0 {
			return
		}
		doc.rawAuthors = v
	}

	var (
		rawAuthors    = strings.Split(doc.rawAuthors, `;`)
		authorKey     = DocAttrAuthor
		emailKey      = docAttrEmail
		initialsKey   = docAttrAuthorInitials
		firstNameKey  = docAttrFirstName
		middleNameKey = docAttrMiddleName
		lastNameKey   = docAttrLastName

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
			sb.WriteString(`, `)
		}
		sb.WriteString(author.FullName())

		if x == 0 {
			doc.Attributes.Entry[authorKey] = author.FullName()
			doc.Attributes.Entry[emailKey] = author.Email
			doc.Attributes.Entry[initialsKey] = author.Initials
			doc.Attributes.Entry[firstNameKey] = author.FirstName
			doc.Attributes.Entry[middleNameKey] = author.MiddleName
			doc.Attributes.Entry[lastNameKey] = author.LastName

			// No continue, the first author have two keys, one is
			// `author` and another is `author_1`.
		}

		authorKey = fmt.Sprintf(`%s_%d`, DocAttrAuthor, x+1)
		emailKey = fmt.Sprintf(`%s_%d`, docAttrEmail, x+1)
		initialsKey = fmt.Sprintf(`%s_%d`, docAttrAuthorInitials, x+1)
		firstNameKey = fmt.Sprintf(`%s_%d`, docAttrFirstName, x+1)
		middleNameKey = fmt.Sprintf(`%s_%d`, docAttrMiddleName, x+1)
		lastNameKey = fmt.Sprintf(`%s_%d`, docAttrLastName, x+1)

		doc.Attributes.Entry[authorKey] = author.FullName()
		doc.Attributes.Entry[emailKey] = author.Email
		doc.Attributes.Entry[initialsKey] = author.Initials
		doc.Attributes.Entry[firstNameKey] = author.FirstName
		doc.Attributes.Entry[middleNameKey] = author.MiddleName
		doc.Attributes.Entry[lastNameKey] = author.LastName
	}

	v = sb.String()
	if len(v) > 0 {
		doc.Attributes.Entry[DocAttrAuthorNames] = v
	}
}

func (doc *Document) unpackRawRevision() {
	if len(doc.rawRevision) > 0 {
		doc.Revision = parseRevision(doc.rawRevision)
		doc.Attributes.Entry[docAttrRevNumber] = doc.Revision.Number
		doc.Attributes.Entry[docAttrRevDate] = doc.Revision.Date
		doc.Attributes.Entry[docAttrRevRemark] = doc.Revision.Remark
		return
	}
	doc.Revision.Number = doc.Attributes.Entry[docAttrRevNumber]
	doc.Revision.Date = doc.Attributes.Entry[docAttrRevDate]
	doc.Revision.Remark = doc.Attributes.Entry[docAttrRevRemark]
}

func (doc *Document) unpackRawTitle() {
	var (
		title string
		prev  byte
		x     int
	)

	if len(doc.Title.raw) == 0 {
		doc.Title.raw = doc.Attributes.Entry[docAttrDocTitle]
		if len(doc.Title.raw) == 0 {
			doc.Title.raw = doc.Attributes.Entry[docAttrTitle]
			if len(doc.Title.raw) == 0 {
				return
			}
		}
	}

	doc.Title.el = parseInlineMarkup(doc, []byte(doc.Title.raw))
	title = doc.Title.el.toText()
	doc.Attributes.Entry[docAttrDocTitle] = title

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
// meta attribute `title-separator` value.
func (doc *Document) unpackTitleSeparator() {
	var (
		v  string
		ok bool
	)

	v, ok = doc.Attributes.Entry[docAttrTitleSeparator]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			doc.Title.sep = v[0]
		}
	}
}
