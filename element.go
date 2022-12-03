// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
)

// element is the building block of asciidoc document.
type element struct {
	// title is the parsed rawTitle for section L1 or parsed raw for
	// section L2-L5.
	title *element

	prev   *element
	next   *element
	parent *element
	child  *element
	label  *element

	table *elementTable

	// sectnums contain the current section numbers.
	// It will be set only if attribute `sectnums` is on.
	sectnums *sectionCounters

	// The key and value for attribute (lineKindAttribute).
	key   string
	value string

	rawTitle string
	Text     string // The content of element without inline formatting.

	raw []byte // Unparsed content of element.

	elementAttribute

	rawLabel       bytes.Buffer
	level          int // The number of dot for ordered list, or '*' for unordered list.
	listItemNumber int // The counter for list item, start from 1.
	kind           int

	// List of substitutions to be applied on raw.
	applySubs int
}

func (el *element) getListOrderedClass() string {
	switch el.level {
	case 2:
		return classNameLoweralpha
	case 3:
		return classNameLowerroman
	case 4:
		return classNameUpperalpha
	case 5:
		return classNameUpperroman
	}
	return classNameArabic
}

func (el *element) getListOrderedType() string {
	switch el.level {
	case 2:
		return `a`
	case 3:
		return `i`
	case 4:
		return `A`
	case 5:
		return `I`
	}
	return ``
}

// getVideoSource generate video full URL for HTML attribute `src`.
func (el *element) getVideoSource() string {
	var (
		u          = new(url.URL)
		src string = el.Attrs[attrNameSrc]

		q        []string
		fragment string
		vstr     string

		isYoutube bool
		isVimeo   bool
		ok        bool
	)

	if el.rawStyle == attrNameYoutube {
		isYoutube = true
	}
	if el.rawStyle == attrNameVimeo {
		isVimeo = true
	}

	if isYoutube {
		u.Scheme = `https`
		u.Host = `www.youtube.com`
		u.Path = `/embed/` + src

		q = append(q, `rel=0`)

		vstr, ok = el.Attrs[attrNameStart]
		if ok {
			q = append(q, attrNameStart+`=`+vstr)
		}

		vstr, ok = el.Attrs[attrNameEnd]
		if ok {
			q = append(q, attrNameEnd+`=`+vstr)
		}

		for _, vstr = range el.options {
			switch vstr {
			case optNameAutoplay, optNameLoop:
				q = append(q, vstr+`=1`)
			case optVideoModest:
				q = append(q, optVideoYoutubeModestbranding+`=1`)
			case optNameNocontrols:
				q = append(q, optNameControls+`=0`)
				q = append(q, optVideoPlaylist+`=`+src)
			case optVideoNofullscreen:
				q = append(q, optVideoFullscreen+`=0`)
				el.Attrs[optVideoNofullscreen] = ``
			}
		}

		vstr, ok = el.Attrs[attrNameTheme]
		if ok {
			q = append(q, attrNameTheme+`=`+vstr)
		}

		vstr, ok = el.Attrs[attrNameLang]
		if ok {
			q = append(q, attrNameYoutubeLang+`=`+vstr)
		}

	} else if isVimeo {
		u.Scheme = `https`
		u.Host = `player.vimeo.com`
		u.Path = `/video/` + src

		for _, vstr = range el.options {
			switch vstr {
			case optNameAutoplay, optNameLoop:
				q = append(q, vstr+`=1`)
			}
		}
		vstr, ok = el.Attrs[attrNameStart]
		if ok {
			fragment = `at=` + vstr
		}

	} else {
		for _, vstr = range el.options {
			switch vstr {
			case optNameAutoplay, optNameLoop:
				el.Attrs[optNameNocontrols] = ``
				el.Attrs[vstr] = ``
			}
		}

		vstr, ok = el.Attrs[attrNameStart]
		if ok {
			fragment = `t=` + vstr
			vstr, ok = el.Attrs[attrNameEnd]
			if ok {
				fragment += `,` + vstr
			}
		} else if vstr, ok = el.Attrs[attrNameEnd]; ok {
			fragment = `t=0,` + vstr
		}

		if len(fragment) > 0 {
			src = src + `#` + fragment
		}
		return src
	}
	u.RawQuery = strings.Join(q, `&amp;`)
	u.Fragment = fragment

	return u.String()
}

func (el *element) hasStyle(s int64) bool {
	return el.style&s > 0
}

func (el *element) isStyleAdmonition() bool {
	return isStyleAdmonition(el.style)
}

func (el *element) isStyleHorizontal() bool {
	return el.style&styleDescriptionHorizontal > 0
}

func (el *element) isStyleQandA() bool {
	return el.style&styleDescriptionQandA > 0
}

func (el *element) isStyleQuote() bool {
	return isStyleQuote(el.style)
}

func (el *element) isStyleVerse() bool {
	return isStyleVerse(el.style)
}

func (el *element) Write(b []byte) {
	el.raw = append(el.raw, b...)
}

func (el *element) WriteByte(b byte) {
	el.raw = append(el.raw, b)
}

func (el *element) WriteString(s string) {
	el.raw = append(el.raw, []byte(s)...)
}

// addChild push the `child` to the list of current element's child.
func (el *element) addChild(child *element) {
	if child == nil {
		return
	}

	child.parent = el
	child.next = nil
	child.prev = nil

	if el.child == nil {
		el.child = child
	} else {
		var c *element = el.child
		for c.next != nil {
			c = c.next
		}
		c.next = child
		child.prev = c
	}
}

// backTrimSpace remove trailing white spaces on raw field.
func (el *element) backTrimSpace() {
	var x int = len(el.raw) - 1
	for ; x > 0; x-- {
		if ascii.IsSpace(el.raw[x]) {
			continue
		}
		break
	}
	el.raw = el.raw[:x+1]
}

func (el *element) lastSuccessor() (last *element) {
	if el.child == nil {
		return nil
	}
	last = el
	for last.child != nil {
		last = last.child
		for last.next != nil {
			last = last.next
		}
	}
	return last
}

func (el *element) parseBlockAudio(doc *Document, line []byte) bool {
	line = bytes.TrimRight(line[7:], " \t")

	var attrBegin int = bytes.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}

	var attrEnd int = bytes.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	var src []byte = bytes.TrimRight(line[:attrBegin], " \t")
	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}
	el.parseElementAttribute(line[attrBegin : attrEnd+1])

	src = applySubstitutions(doc, []byte(src))
	el.Attrs[attrNameSrc] = string(src)

	return true
}

// parseBlockImage parse the image block or line.
// The line parameter must not have "image::" block or "image:" macro prefix.
func (el *element) parseBlockImage(doc *Document, line []byte) bool {
	var (
		attrBegin int = bytes.IndexByte(line, '[')

		attr string
		key  string
		val  string
		kv   []string

		attrs [][]byte
		src   []byte
		battr []byte
		alt   []byte

		attrEnd int
		dot     int
		x       int
	)

	if attrBegin < 0 {
		return false
	}
	attrEnd = bytes.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	src = bytes.TrimRight(line[:attrBegin], " \t")

	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}
	src = applySubstitutions(doc, src)
	el.Attrs[attrNameSrc] = string(src)

	attrs = bytes.Split(line[attrBegin+1:attrEnd], []byte(`,`))
	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}
	var hasWidth bool
	for x, battr = range attrs {
		attr = string(battr)
		if x == 0 {
			alt = bytes.TrimSpace(attrs[0])
			if len(alt) == 0 {
				dot = bytes.IndexByte(src, '.')
				if dot > 0 {
					alt = src[:dot]
				}
			}
			el.Attrs[attrNameAlt] = string(alt)
			continue
		}
		if x == 1 {
			if ascii.IsDigits(attrs[1]) {
				el.Attrs[attrNameWidth] = string(attrs[1])
				hasWidth = true
				continue
			}
		}
		if hasWidth && x == 2 {
			if ascii.IsDigits(attrs[2]) {
				el.Attrs[attrNameHeight] = string(attrs[2])
			}
		}
		kv = strings.SplitN(attr, `=`, 2)
		if len(kv) != 2 {
			continue
		}
		key = strings.TrimSpace(kv[0])
		val = strings.Trim(kv[1], `"`)
		switch key {
		case attrNameFloat, attrNameAlign, attrNameRole:
			if val == `center` {
				val = `text-center`
			}
			el.addRole(val)
		default:
			el.Attrs[key] = val
		}
	}

	for key, val = range el.Attrs {
		if key == attrNameLink {
			val = string(applySubstitutions(doc, []byte(val)))
			el.Attrs[key] = val
		}
	}

	return true
}

func (el *element) parseInlineMarkup(doc *Document, kind int) {
	if len(el.raw) == 0 {
		return
	}

	var container *element = parseInlineMarkup(doc, el.raw)
	if kind != 0 {
		container.kind = kind
	}
	container.parent = el
	container.next = el.child
	if el.child != nil {
		el.child.prev = container
	}
	el.child = container

	el.raw = nil
}

func (el *element) parseLineAdmonition(line []byte) {
	var (
		sep      int    = bytes.IndexByte(line, ':')
		class    string = string(bytes.ToLower(line[:sep]))
		rawLabel string = admonitionToLabel(class)
	)

	el.addRole(class)
	el.rawLabel.WriteString(rawLabel)

	line = bytes.TrimSpace(line[sep+1:])
	el.Write(line)
	el.WriteByte('\n')
}

func (el *element) parseListDescriptionItem(line []byte) {
	var (
		label []byte
		x     int
		c     byte
	)

	label, x = indexUnescape(line, []byte(`::`))
	el.rawLabel.Write(label)

	line = line[x+2:]
	for x, c = range line {
		if c == ':' {
			el.level++
			continue
		}
		break
	}

	// Skip leading spaces...
	if x < len(line)-1 {
		line = line[x:]
	} else {
		line = nil
	}
	for x, c = range line {
		if c == ' ' || c == '\t' {
			continue
		}
		break
	}
	if len(line) > 0 {
		el.Write(line[x:])
		el.WriteByte('\n')
	}
}

func (el *element) parseListOrderedItem(line []byte) {
	var (
		x int
	)

	for ; x < len(line); x++ {
		if line[x] == '.' {
			el.level++
			continue
		}
		if line[x] == ' ' || line[x] == '\t' {
			break
		}
	}
	for ; x < len(line); x++ {
		if line[x] == ' ' || line[x] == '\t' {
			continue
		}
		break
	}
	el.Write(line[x:])
	el.WriteByte('\n')
}

func (el *element) parseListUnorderedItem(line []byte) {
	var (
		x int
	)

	for ; x < len(line); x++ {
		if line[x] == '*' {
			el.level++
			continue
		}
		if line[x] == ' ' || line[x] == '\t' {
			break
		}
	}
	for ; x < len(line); x++ {
		if line[x] == ' ' || line[x] == '\t' {
			continue
		}
		break
	}
	if len(line[x:]) > 3 {
		var (
			checklist = line[x : x+3]
			sym       string
		)
		if bytes.Equal(checklist, []byte(`[ ]`)) {
			sym = symbolUnchecked
		} else if bytes.Equal(checklist, []byte(`[x]`)) ||
			bytes.Equal(checklist, []byte(`[X]`)) ||
			bytes.Equal(checklist, []byte(`[*]`)) {
			sym = symbolChecked
		}
		if len(sym) != 0 {
			el.WriteString(sym)
			el.WriteByte(' ')
			el.addRole(classNameChecklist)
			x += 3
			for ; x < len(line); x++ {
				if line[x] == ' ' || line[x] == '\t' {
					continue
				}
				break
			}

		}
	}
	el.Write(line[x:])
	el.WriteByte('\n')
}

func (el *element) parseSection(doc *Document, isDiscrete bool) {
	if !isDiscrete {
		el.level = (el.kind - elKindSectionL1) + 1
	}

	var (
		container *element = parseInlineMarkup(doc, el.raw)

		lastChild *element
		p         *element
		anc       *anchor

		refText string
		ok      bool
	)

	if len(el.ID) == 0 {
		lastChild = container.lastSuccessor()
		if lastChild != nil && lastChild.kind == elKindInlineID {
			el.ID = lastChild.ID

			// Delete last child
			if lastChild.prev != nil {
				p = lastChild.prev
				p.next = nil
			} else if lastChild.parent != nil {
				p = lastChild.parent
				p.child = nil
			}
			lastChild.prev = nil
			lastChild.parent = nil
		}
	}

	container.parent = el
	el.title = container
	el.raw = nil
	el.Text = container.toText()

	if len(el.ID) == 0 {
		_, ok = doc.Attributes[metaNameSectIDs]
		if ok {
			el.ID = generateID(doc, el.Text)
			el.ID = doc.registerAnchor(el.ID, el.Text)
		}
	}

	refText, ok = el.Attrs[attrNameRefText]
	if ok {
		doc.titleID[refText] = el.ID
		// Replace the label with refText.
		anc = doc.anchors[el.ID]
		if anc != nil {
			anc.label = refText
		}
	}
	doc.titleID[el.Text] = el.ID

	_, ok = doc.Attributes[metaNameSectNums]
	if ok && !isDiscrete {
		el.sectnums = doc.sectnums.set(el.level)
	}
}

func (el *element) parseStyleClass(line []byte) {
	line = bytes.Trim(line, `[]`)

	var (
		parts = bytes.Split(line, []byte(`.`))

		class []byte
	)

	for _, class = range parts {
		class = bytes.TrimSpace(class)
		if len(class) > 0 {
			el.addRole(string(class))
		}
	}
}

func (el *element) parseBlockVideo(doc *Document, line []byte) bool {
	line = bytes.TrimRight(line[7:], " \t")

	var (
		attrBegin = bytes.IndexByte(line, '[')

		videoSrc []byte
		attrEnd  int
	)

	if attrBegin < 0 {
		return false
	}
	attrEnd = bytes.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}

	videoSrc = bytes.TrimRight(line[:attrBegin], " \t")
	videoSrc = applySubstitutions(doc, []byte(videoSrc))
	el.Attrs[attrNameSrc] = string(videoSrc)

	el.parseElementAttribute(line[attrBegin : attrEnd+1])

	return true
}

func (el *element) postParseList(doc *Document, kind int) {
	var (
		item *element = el.child
		raw  []byte
	)

	for item != nil {
		if item.kind == kind {
			if item.kind == elKindListDescriptionItem {
				raw = item.rawLabel.Bytes()
				item.label = parseInlineMarkup(doc, raw)
				item.rawLabel.Reset()
			}
			item.parseInlineMarkup(doc, elKindInlineParagraph)
		}
		item = item.next
	}
}

// postParseParagraph check if paragraph is a blockquote based on the first
// character of the first line ('"'), the last character of last second line
// ('"'), and the last line start with "-- ".
func (el *element) postParseParagraph(parent *element) {
	if el.isStyleQuote() {
		return
	}
	if parent != nil && parent.kind == elKindBlockExcerpts {
		return
	}

	el.raw = bytes.TrimRight(el.raw, " \t\n")

	var (
		lines [][]byte = bytes.Split(el.raw, []byte{'\n'})
	)

	if len(lines) <= 1 {
		return
	}

	el.postParseParagraphAsQuote(lines)
}

func (el *element) postParseParagraphAsQuote(lines [][]byte) bool {
	var (
		lastLine []byte = lines[len(lines)-1]

		firstLine      []byte
		secondLastLine []byte
		line           []byte

		secondLastIdx int
		x             int
	)

	if len(lastLine) <= 4 {
		return false
	}
	if lastLine[0] != '-' {
		return false
	}
	if lastLine[1] != '-' {
		return false
	}
	if !(lastLine[2] == ' ' || lastLine[2] == '\t') {
		return false
	}
	firstLine = lines[0]
	if firstLine[0] != '"' {
		return false
	}
	secondLastLine = lines[len(lines)-2]
	if secondLastLine[len(secondLastLine)-1] != '"' {
		return false
	}

	el.raw = el.raw[:0]

	secondLastIdx = len(lines) - 2

	for x, line = range lines[:len(lines)-1] {
		if x == 0 {
			if x == secondLastIdx {
				el.Write(line[1 : len(line)-1])
			} else {
				el.Write(line[1:])
			}
		} else if x == secondLastIdx {
			el.Write(line[:len(line)-1])
		} else {
			el.Write(line)
		}
		el.WriteByte('\n')
	}

	el.kind = elKindBlockExcerpts
	var opts []string = strings.SplitN(string(lastLine[3:]), `,`, 2)
	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}
	if len(opts) >= 1 {
		el.Attrs[attrNameAttribution] = strings.TrimSpace(opts[0])
	}
	if len(opts) >= 2 {
		el.Attrs[attrNameCitation] = strings.TrimSpace(opts[1])
	}

	return true
}

// postConsumeTable after we get all raw tables contents, we split them into
// multiple rows, based on empty line between row.
func (el *element) postConsumeTable() (table *elementTable) {
	el.table = newTable(&el.elementAttribute, el.raw)
	return el.table
}

func (el *element) removeLastIfEmpty() {
	if el.child == nil {
		return
	}
	var c *element = el
	for c.child != nil {
		c = c.child
		for c.next != nil {
			c = c.next
		}
	}
	if c.kind != elKindText || len(c.raw) > 0 {
		return
	}
	if c.prev != nil {
		c.prev.next = nil
		if c.prev.kind == elKindText {
			el.raw = bytes.TrimRight(el.raw, " \t")
		}
	} else if c.parent != nil {
		c.parent.child = nil
	}
	c.prev = nil
	c.parent = nil
}

func (el *element) setStyleAdmonition(admName string) {
	admName = strings.ToLower(admName)
	el.addRole(admName)
	var rawLabel string = admonitionToLabel(admName)
	el.rawLabel.WriteString(rawLabel)
}

func (el *element) toHTML(doc *Document, w io.Writer) {
	switch el.kind {
	case lineKindAttribute:
		doc.Attributes.apply(el.key, el.value)

	case elKindCrossReference:
		var (
			href   string  = el.Attrs[attrNameHref]
			label          = string(el.raw)
			anchor *anchor = doc.anchors[href]
		)
		if anchor == nil {
			href = doc.titleID[href]
			if len(href) > 0 {
				anchor = doc.anchors[href]
				if anchor != nil {
					if len(label) == 0 {
						label = anchor.label
					}
				}
			} else {
				// href is not ID nor label, assume its broken
				// link.
				href = el.Attrs[attrNameHref]
				if len(label) == 0 {
					label = href
				}
			}
		} else {
			if len(label) == 0 {
				label = anchor.label
			}
		}
		fmt.Fprintf(w, `<a href="#%s">%s</a>`, href, label)

	case elKindFootnote:
		htmlWriteFootnote(el, w)

	case elKindMacroTOC:
		if doc.tocIsEnabled && doc.tocPosition == metaValueMacro {
			doc.tocHTML(w)
		}

	case elKindPreamble:
		if !doc.isEmbedded {
			fmt.Fprint(w, _htmlPreambleBegin)
		}

	case elKindSectionDiscrete:
		hmltWriteSectionDiscrete(doc, el, w)

	case elKindSectionL1, elKindSectionL2, elKindSectionL3,
		elKindSectionL4, elKindSectionL5:
		htmlWriteSection(doc, el, w)

	case elKindParagraph:
		if el.isStyleAdmonition() {
			htmlWriteBlockAdmonition(el, w)
		} else if el.isStyleQuote() {
			htmlWriteBlockQuote(el, w)
		} else if el.isStyleVerse() {
			htmlWriteBlockVerse(el, w)
		} else {
			htmlWriteParagraphBegin(el, w)
		}

	case elKindLiteralParagraph, elKindBlockLiteral,
		elKindBlockLiteralNamed,
		elKindBlockListing, elKindBlockListingNamed:
		htmlWriteBlockLiteral(el, w)

	case elKindInlineImage:
		htmlWriteInlineImage(el, w)

	case elKindInlinePass:
		htmlWriteInlinePass(doc, el, w)

	case elKindListDescription:
		htmlWriteListDescription(el, w)
	case elKindListOrdered:
		htmlWriteListOrdered(el, w)
	case elKindListUnordered:
		htmlWriteListUnordered(el, w)

	case elKindListOrderedItem, elKindListUnorderedItem:
		fmt.Fprint(w, "\n<li>")

	case elKindListDescriptionItem:
		var (
			format string
			label  bytes.Buffer
		)
		if el.label != nil {
			el.label.toHTML(doc, &label)
		} else {
			label.Write(el.rawLabel.Bytes())
		}

		if el.isStyleQandA() {
			format = _htmlListDescriptionItemQandABegin
		} else if el.isStyleHorizontal() {
			format = _htmlListDescriptionItemHorizontalBegin
		} else {
			format = _htmlListDescriptionItemBegin
		}
		fmt.Fprintf(w, format, label.String())

	case lineKindHorizontalRule:
		fmt.Fprint(w, "\n<hr>")

	case lineKindPageBreak:
		fmt.Fprint(w, "\n<div style=\"page-break-after: always;\"></div>")

	case elKindBlockExample:
		if el.isStyleAdmonition() {
			htmlWriteBlockAdmonition(el, w)
		} else {
			htmlWriteBlockExample(doc, el, w)
		}

	case elKindBlockImage:
		htmlWriteBlockImage(doc, el, w)

	case elKindBlockOpen:
		if el.isStyleAdmonition() {
			htmlWriteBlockAdmonition(el, w)
		} else if el.isStyleQuote() {
			htmlWriteBlockQuote(el, w)
		} else if el.isStyleVerse() {
			htmlWriteBlockVerse(el, w)
		} else {
			htmlWriteBlockOpenBegin(el, w)
		}

	case elKindBlockPassthrough:
		fmt.Fprintf(w, "\n%s", el.raw)

	case elKindBlockExcerpts:
		if el.isStyleVerse() {
			htmlWriteBlockVerse(el, w)
		} else {
			htmlWriteBlockQuote(el, w)
		}

	case elKindBlockSidebar:
		htmlWriteBlockSidebar(el, w)

	case elKindBlockVideo:
		htmlWriteBlockVideo(el, w)

	case elKindBlockAudio:
		htmlWriteBlockAudio(el, w)

	case elKindInlineID:
		if !doc.isForToC {
			fmt.Fprintf(w, "<a id=%q></a>", el.ID)
		}

	case elKindInlineIDShort:
		if !doc.isForToC {
			fmt.Fprintf(w, "<span id=%q>%s", el.ID, el.raw)
		}

	case elKindInlineParagraph:
		fmt.Fprintf(w, "\n<p>%s", el.raw)

	case elKindPassthrough:
		fmt.Fprint(w, string(el.raw))
	case elKindPassthroughDouble:
		fmt.Fprint(w, string(el.raw))
	case elKindPassthroughTriple:
		fmt.Fprint(w, string(el.raw))

	case elKindSymbolQuoteDoubleBegin:
		fmt.Fprint(w, symbolQuoteDoubleBegin, string(el.raw))
	case elKindSymbolQuoteDoubleEnd:
		fmt.Fprint(w, symbolQuoteDoubleEnd, string(el.raw))

	case elKindSymbolQuoteSingleBegin:
		fmt.Fprint(w, symbolQuoteSingleBegin, string(el.raw))
	case elKindSymbolQuoteSingleEnd:
		fmt.Fprint(w, symbolQuoteSingleEnd, string(el.raw))

	case elKindText:
		fmt.Fprint(w, string(el.raw))

	case elKindTextBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(w, "<strong>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "*")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(w, "<strong>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "**")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindTextItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(w, "<em>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "_")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(w, "<em>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "__")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindTextMono:
		if el.hasStyle(styleTextMono) {
			fmt.Fprint(w, "<code>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "`")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedMono:
		if el.hasStyle(styleTextMono) {
			fmt.Fprint(w, "<code>")
		} else if len(el.raw) > 0 {
			fmt.Fprint(w, "``")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindURL:
		htmlWriteURLBegin(el, w)

	case elKindTextSubscript:
		fmt.Fprintf(w, "<sub>%s</sub>", el.raw)
	case elKindTextSuperscript:
		fmt.Fprintf(w, "<sup>%s</sup>", el.raw)

	case elKindTable:
		htmlWriteTable(doc, el, w)
	}

	if el.child != nil {
		el.child.toHTML(doc, w)
	}

	switch el.kind {
	case elKindPreamble:
		if !doc.isEmbedded {
			fmt.Fprint(w, "\n</div>")
		}
		if doc.tocIsEnabled && doc.tocPosition == metaValuePreamble {
			doc.tocHTML(w)
		}
		if !doc.isEmbedded {
			fmt.Fprint(w, "\n</div>")
		}

	case elKindSectionL1, elKindSectionL2, elKindSectionL3,
		elKindSectionL4, elKindSectionL5:
		if el.kind == elKindSectionL1 {
			fmt.Fprint(w, "\n</div>")
		}
		fmt.Fprint(w, "\n</div>")

	case elKindParagraph:
		if el.isStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if el.isStyleQuote() {
			htmlWriteBlockQuoteEnd(el, w)
		} else if el.isStyleVerse() {
			htmlWriteBlockVerseEnd(el, w)
		} else {
			fmt.Fprint(w, "</p>\n</div>")
		}

	case elKindListOrderedItem, elKindListUnorderedItem:
		fmt.Fprint(w, "\n</li>")

	case elKindListDescriptionItem:
		var format string
		if el.isStyleQandA() {
			format = "\n</li>"
		} else if el.isStyleHorizontal() {
			format = "\n</td>\n</tr>"
		} else {
			format = "\n</dd>"
		}
		fmt.Fprint(w, format)

	case elKindListDescription:
		htmlWriteListDescriptionEnd(el, w)
	case elKindListOrdered:
		htmlWriteListOrderedEnd(w)
	case elKindListUnordered:
		htmlWriteListUnorderedEnd(w)

	case elKindBlockExample:
		if el.isStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else {
			fmt.Fprint(w, "\n</div>\n</div>")
		}

	case elKindBlockOpen:
		if el.isStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if el.isStyleQuote() {
			htmlWriteBlockQuoteEnd(el, w)
		} else if el.isStyleVerse() {
			htmlWriteBlockVerseEnd(el, w)
		} else {
			fmt.Fprint(w, "\n</div>\n</div>")
		}
	case elKindBlockExcerpts:
		if el.isStyleVerse() {
			htmlWriteBlockVerseEnd(el, w)
		} else {
			htmlWriteBlockQuoteEnd(el, w)
		}

	case elKindBlockSidebar:
		fmt.Fprint(w, "\n</div>\n</div>")

	case elKindInlineIDShort:
		if !doc.isForToC {
			fmt.Fprint(w, "</span>")
		}

	case elKindInlineParagraph:
		fmt.Fprint(w, "</p>")

	case elKindTextBold, elKindUnconstrainedBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(w, "</strong>")
		}
	case elKindTextItalic, elKindUnconstrainedItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(w, "</em>")
		}
	case elKindTextMono, elKindUnconstrainedMono:
		if el.hasStyle(styleTextMono) {
			fmt.Fprint(w, "</code>")
		}
	case elKindURL:
		htmlWriteURLEnd(w)
	}

	if el.next != nil {
		el.next.toHTML(doc, w)
	}
}

func (el *element) toText() (text string) {
	var buf bytes.Buffer
	el.writeText(&buf)
	return buf.String()
}

func (el *element) writeText(w io.Writer) {
	switch el.kind {
	case elKindPassthrough:
		fmt.Fprint(w, string(el.raw))
	case elKindPassthroughDouble:
		fmt.Fprint(w, string(el.raw))
	case elKindPassthroughTriple:
		fmt.Fprint(w, string(el.raw))

	case elKindSymbolQuoteDoubleBegin:
		fmt.Fprint(w, symbolQuoteDoubleBegin, string(el.raw))

	case elKindSymbolQuoteDoubleEnd:
		fmt.Fprint(w, symbolQuoteDoubleEnd, string(el.raw))

	case elKindSymbolQuoteSingleBegin:
		fmt.Fprint(w, symbolQuoteSingleBegin, string(el.raw))
	case elKindSymbolQuoteSingleEnd:
		fmt.Fprint(w, symbolQuoteSingleEnd, string(el.raw))

	case elKindText:
		fmt.Fprint(w, string(el.raw))

	case elKindTextBold:
		if !el.hasStyle(styleTextBold) {
			fmt.Fprint(w, "*")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedBold:
		if !el.hasStyle(styleTextBold) {
			fmt.Fprint(w, "**")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindTextItalic:
		if !el.hasStyle(styleTextItalic) {
			fmt.Fprint(w, "_")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedItalic:
		if !el.hasStyle(styleTextItalic) {
			fmt.Fprint(w, "__")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindTextMono:
		if !el.hasStyle(styleTextMono) {
			fmt.Fprint(w, "`")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindUnconstrainedMono:
		if !el.hasStyle(styleTextMono) {
			fmt.Fprint(w, "``")
		}
		fmt.Fprint(w, string(el.raw))

	case elKindURL:
		fmt.Fprint(w, string(el.raw))
	case elKindTextSubscript:
		fmt.Fprint(w, string(el.raw))
	case elKindTextSuperscript:
		fmt.Fprint(w, string(el.raw))
	}

	if el.child != nil {
		el.child.writeText(w)
	}
	if el.next != nil {
		el.next.writeText(w)
	}
}

func admonitionToLabel(admName string) string {
	admName = strings.ToUpper(admName)
	switch admName {
	case admonitionCaution:
		return "Caution"
	case admonitionImportant:
		return "Important"
	case admonitionNote:
		return "Note"
	case admonitionTip:
		return "Tip"
	case admonitionWarning:
		return "Warning"
	}
	return admName
}
