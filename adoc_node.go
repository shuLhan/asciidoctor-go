// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"text/template"

	"github.com/shuLhan/share/lib/ascii"
)

//
// adocNode is the building block of asciidoc document.
//
type adocNode struct {
	ID       string
	Attrs    map[string]string
	Opts     map[string]string
	Text     string // The content of node without inline formatting.
	kind     int
	level    int    // The number of dot for ordered list, or star '*' for unordered list.
	raw      []byte // unparsed content of node.
	rawLabel bytes.Buffer
	rawTitle string
	style    int64
	classes  []string

	// The key and value for attribute (lineKindAttribute).
	key   string
	value string

	// title is the parsed rawTitle for section L1 or parsed raw for
	// section L2-L5.
	title *adocNode

	parent *adocNode
	child  *adocNode
	next   *adocNode
	prev   *adocNode
}

func (node *adocNode) Classes() string {
	if len(node.classes) == 0 {
		return ""
	}
	return strings.Join(node.classes, " ")
}

func (node *adocNode) Content() string {
	node.raw = bytes.TrimRight(node.raw, "\n")
	return string(node.raw)
}

func (node *adocNode) GetListOrderedClass() string {
	switch node.level {
	case 2:
		return "loweralpha"
	case 3:
		return "lowerroman"
	case 4:
		return "upperalpha"
	case 5:
		return "upperroman"
	}
	return "arabic"
}

func (node *adocNode) GetListOrderedType() string {
	switch node.level {
	case 2:
		return "a"
	case 3:
		return "i"
	case 4:
		return "A"
	case 5:
		return "I"
	}
	return ""
}

//
// GetVideoSource generate video full URL for HTML attribute "src".
//
func (node *adocNode) GetVideoSource() string {
	var (
		u        = new(url.URL)
		q        []string
		fragment string
	)

	src := node.Attrs[attrNameSrc]
	opts := strings.Split(node.Attrs[attrNameOptions], ",")

	_, ok := node.Attrs[attrNameYoutube]
	if ok {
		u.Scheme = "https"
		u.Host = "www.youtube.com"
		u.Path = "/embed/" + src

		q = append(q, "rel=0")

		v, ok := node.Attrs[attrNameStart]
		if ok {
			q = append(q, attrNameStart+"="+v)
		}
		v, ok = node.Attrs[attrNameEnd]
		if ok {
			q = append(q, attrNameEnd+"="+v)
		}
		for _, opt := range opts {
			opt = strings.TrimSpace(opt)
			switch opt {
			case optNameAutoplay, optNameLoop:
				q = append(q, opt+"=1")
			case optVideoModest:
				q = append(q, optVideoYoutubeModestbranding+"=1")
			case optNameNocontrols:
				q = append(q, optNameControls+"=0")
				q = append(q, optVideoPlaylist+"="+src)
			case optVideoNofullscreen:
				q = append(q, optVideoFullscreen+"=0")
				node.Attrs[optVideoNofullscreen] = "1"
			}
		}
		v, ok = node.Attrs[attrNameTheme]
		if ok {
			q = append(q, attrNameTheme+"="+v)
		}
		v, ok = node.Attrs[attrNameLang]
		if ok {
			q = append(q, attrNameYoutubeLang+"="+v)
		}

	} else if _, ok = node.Attrs[attrNameVimeo]; ok {
		u.Scheme = "https"
		u.Host = "player.vimeo.com"
		u.Path = "/video/" + src

		for _, opt := range opts {
			opt = strings.TrimSpace(opt)
			switch opt {
			case optNameAutoplay, optNameLoop:
				q = append(q, opt+"=1")
			}
		}
		v, ok := node.Attrs[attrNameStart]
		if ok {
			fragment = "at=" + v
		}
	} else {
		for _, opt := range opts {
			opt = strings.TrimSpace(opt)
			switch opt {
			case optNameAutoplay, optNameLoop:
				node.Attrs[optNameNocontrols] = "1"
				node.Attrs[opt] = "1"
			}
		}

		v, ok := node.Attrs[attrNameStart]
		if ok {
			fragment = "t=" + v
			v, ok = node.Attrs[attrNameEnd]
			if ok {
				fragment += "," + v
			}
		} else if v, ok = node.Attrs[attrNameEnd]; ok {
			fragment = "t=0," + v
		}

		if len(fragment) > 0 {
			src = src + "#" + fragment
		}
		return src
	}
	u.RawQuery = strings.Join(q, "&amp;")
	u.Fragment = fragment

	return u.String()
}

func (node *adocNode) HasStyle(s int64) bool {
	return node.style&s > 0
}

func (node *adocNode) IsStyleAdmonition() bool {
	return isStyleAdmonition(node.style)
}

func (node *adocNode) IsStyleHorizontal() bool {
	return node.style&styleDescriptionHorizontal > 0
}

func (node *adocNode) IsStyleListing() bool {
	return node.style&styleBlockListing > 0
}

func (node *adocNode) IsStyleQandA() bool {
	return node.style&styleDescriptionQandA > 0
}

func (node *adocNode) IsStyleQuote() bool {
	return isStyleQuote(node.style)
}

func (node *adocNode) IsStyleVerse() bool {
	return isStyleVerse(node.style)
}

func (node *adocNode) Label() string {
	return node.rawLabel.String()
}

func (node *adocNode) QuoteAuthor() string {
	return node.key
}

func (node *adocNode) QuoteCitation() string {
	return node.value
}

func (node *adocNode) Title() string {
	return node.rawTitle
}

func (node *adocNode) URLTarget() string {
	return node.value
}

func (node *adocNode) Write(b []byte) {
	node.raw = append(node.raw, b...)
}

func (node *adocNode) WriteByte(b byte) {
	node.raw = append(node.raw, b)
}

func (node *adocNode) WriteString(s string) {
	node.raw = append(node.raw, []byte(s)...)
}

//
// addChild push the node "child" to the list of current node child.
//
// This function trigger the text substitution on child.
//
func (node *adocNode) addChild(child *adocNode) {
	if child == nil {
		return
	}

	child.parent = node
	child.next = nil
	child.prev = nil

	if node.child == nil {
		node.child = child
	} else {
		c := node.child
		for c.next != nil {
			c = c.next
		}
		c.next = child
		child.prev = c
	}
}

func (node *adocNode) addNext(next *adocNode) {
	if next == nil {
		return
	}
	next.parent = node.parent
	next.prev = node
	node.next = next
}

func (node *adocNode) applySubstitutions() {
	if len(node.rawTitle) > 0 {
		node.rawTitle = htmlSubstituteSpecialChars(node.rawTitle)
	}

	node.raw = bytes.TrimRight(node.raw, "\n")
	content := htmlSubstituteSpecialChars(string(node.raw))

	switch node.kind {
	case nodeKindBlockExample, nodeKindBlockExcerpts, nodeKindParagraph,
		nodeKindBlockSidebar:
		node.raw = []byte(content)
	default:
		node.raw = []byte(content)
	}
}

func (node *adocNode) debug(n int) {
	for x := 0; x < n; x++ {
		fmt.Printf("\t")
	}
	fmt.Printf("node: {kind:%-3d style:%-3d raw:%s}\n", node.kind, node.style, node.raw)
	if node.child != nil {
		node.child.debug(n + 1)
	}
	if node.next != nil {
		node.next.debug(n)
	}
}

func (node *adocNode) lastSuccessor() (last *adocNode) {
	if node.child == nil {
		return nil
	}
	last = node
	for last.child != nil {
		last = last.child
		for last.next != nil {
			last = last.next
		}
	}
	return last
}

func (node *adocNode) parseBlockAudio(line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	src := strings.TrimRight(line[:attrBegin], " \t")
	key, val, attrs := parseAttributeElement(line[attrBegin : attrEnd+1])
	node.Attrs = make(map[string]string, len(attrs)+1)
	node.Attrs[attrNameSrc] = src

	for _, attr := range attrs {
		kv := strings.Split(attr, "=")

		key = strings.ToLower(kv[0])
		if len(kv) >= 2 {
			val = kv[1]
		} else {
			val = "1"
		}

		if key == attrNameOptions {
			node.Attrs[key] = val
			opts := strings.Split(val, ",")
			node.Opts = make(map[string]string, len(opts))
			node.Opts[optNameControls] = "1"

			for _, opt := range opts {
				switch opt {
				case optNameNocontrols:
					node.Opts[optNameControls] = "0"
				case optNameControls:
					node.Opts[optNameControls] = "1"
				default:
					node.Opts[opt] = "1"
				}
			}
		}
	}
	return true
}

//
// parseImage parse the image block or line.
// The line parameter must not have "image::" block or "image:" macro prefix.
//
func (node *adocNode) parseImage(line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	src := strings.TrimRight(line[:attrBegin], " \t")

	if node.Attrs == nil {
		node.Attrs = make(map[string]string)
	}
	node.Attrs[attrNameSrc] = src

	attrs := strings.Split(line[attrBegin+1:attrEnd], ",")
	if node.Attrs == nil {
		node.Attrs = make(map[string]string)
	}
	var hasWidth bool
	for x, attr := range attrs {
		if x == 0 {
			alt := strings.TrimSpace(attrs[0])
			if len(alt) == 0 {
				dot := strings.IndexByte(src, '.')
				if dot > 0 {
					alt = src[:dot]
				}
			}
			node.Attrs[attrNameAlt] = alt
			continue
		}
		if x == 1 {
			if ascii.IsDigits([]byte(attrs[1])) {
				node.Attrs[attrNameWidth] = attrs[1]
				hasWidth = true
				continue
			}
		}
		if hasWidth && x == 2 {
			if ascii.IsDigits([]byte(attrs[2])) {
				node.Attrs[attrNameHeight] = attrs[2]
			}
		}
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) != 2 {
			continue
		}
		val := strings.Trim(kv[1], `"`)
		switch kv[0] {
		case attrNameFloat, attrNameAlign, attrNameRole:
			if val == "center" {
				val = "text-center"
			}
			node.classes = append(node.classes, val)
		}
	}
	return true
}

func (node *adocNode) parseInlineMarkup(doc *Document, kind int) {
	if len(node.raw) == 0 {
		return
	}

	container := parseInlineMarkup(doc, node.raw)
	if kind != 0 {
		container.kind = kind
	}
	container.parent = node
	container.next = node.child
	if node.child != nil {
		node.child.prev = container
	}
	node.child = container

	node.raw = nil
}

func (node *adocNode) parseLineAdmonition(line string) {
	sep := strings.IndexByte(line, ':')
	class := strings.ToLower(line[:sep])
	node.classes = append(node.classes, class)
	node.rawLabel.WriteString(strings.Title(class))
	line = strings.TrimSpace(line[sep+1:])
	node.WriteString(line)
	node.WriteByte('\n')
}

func (node *adocNode) parseListDescriptionItem(line string) {
	var (
		x int
		c rune
	)
	for x, c = range line {
		if c == ':' {
			break
		}
		node.rawLabel.WriteRune(c)
	}
	line = line[x:]
	for x, c = range line {
		if c == ':' {
			node.level++
			continue
		}
		break
	}
	// Skip leading spaces...
	if x < len(line)-1 {
		line = line[x:]
	} else {
		line = ""
	}
	for x, c = range line {
		if c == ' ' || c == '\t' {
			continue
		}
		break
	}
	node.level -= 2
	if len(line) > 0 {
		node.WriteString(line[x:])
	}
}

func (node *adocNode) parseListOrderedItem(line string) {
	x := 0
	for ; x < len(line); x++ {
		if line[x] == '.' {
			node.level++
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
	node.WriteString(line[x:])
	node.WriteByte('\n')
}

func (node *adocNode) parseListUnorderedItem(line string) {
	x := 0
	for ; x < len(line); x++ {
		if line[x] == '*' {
			node.level++
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
	node.WriteString(line[x:])
	node.WriteByte('\n')
}

func (node *adocNode) parseSection(doc *Document) {
	node.level = (node.kind - nodeKindSectionL1) + 1

	container := parseInlineMarkup(doc, node.raw)

	if len(node.ID) == 0 {
		lastChild := container.lastSuccessor()
		if lastChild != nil && lastChild.kind == nodeKindInlineID {
			node.ID = lastChild.ID

			// Delete last child
			if lastChild.prev != nil {
				p := lastChild.prev
				p.next = nil
			} else if lastChild.parent != nil {
				p := lastChild.parent
				p.child = nil
			}
			lastChild.prev = nil
			lastChild.parent = nil
		}
	}

	container.parent = node
	node.title = container
	node.raw = nil

	var text bytes.Buffer
	err := container.toText(&text)
	if err != nil {
		log.Fatalf("parseSection: " + err.Error())
	}
	node.Text = text.String()

	if len(node.ID) == 0 {
		node.ID = generateID(node.Text)
	}
	doc.titleID[node.Text] = node.ID

	refText, ok := node.Attrs[attrNameRefText]
	if ok {
		doc.titleID[refText] = node.ID
	} else {
		refText = node.Text
	}

	doc.registerAnchor(node.ID, refText)
}

func (node *adocNode) parseStyleClass(line string) {
	line = strings.Trim(line, "[]")
	parts := strings.Split(line, ".")
	for _, class := range parts {
		class = strings.TrimSpace(class)
		if len(class) > 0 {
			node.classes = append(node.classes, class)
		}
	}
}

func (node *adocNode) parseVideo(line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	videoSrc := strings.TrimRight(line[:attrBegin], " \t")
	key, val, attrs := parseAttributeElement(line[attrBegin : attrEnd+1])

	if node.Attrs == nil {
		node.Attrs = make(map[string]string, len(attrs)+1)
	}
	node.Attrs[attrNameSrc] = videoSrc

	start := 0
	if key == attrNameYoutube || key == attrNameVimeo {
		node.Attrs[key] = "1"
		start = 1
	}

	for _, attr := range attrs[start:] {
		kv := strings.Split(attr, "=")

		key = strings.ToLower(kv[0])
		if len(kv) >= 2 {
			val = kv[1]
		} else {
			val = "1"
		}

		switch key {
		case attrNameWidth, attrNameHeight,
			attrNameOptions, attrNamePoster, attrNameStart,
			attrNameEnd, attrNameTheme, attrNameLang:
			node.Attrs[key] = val
		}
	}
	return true
}

func (node *adocNode) postParseList(doc *Document, kind int) {
	item := node.child
	for item != nil {
		if item.kind == kind {
			item.parseInlineMarkup(doc, nodeKindInlineParagraph)
		}
		item = item.next
	}
}

//
// postParseParagraph check if paragraph is a blockquote based on the first
// character of the first line ('"'), the last character of last second line
// ('"'), and the last line start with "-- ".
//
func (node *adocNode) postParseParagraph(parent *adocNode) {
	if node.IsStyleQuote() {
		return
	}
	if parent != nil && parent.kind == nodeKindBlockExcerpts {
		return
	}

	node.raw = bytes.TrimRight(node.raw, " \t\n")

	lines := bytes.Split(node.raw, []byte{'\n'})
	if len(lines) <= 1 {
		return
	}

	ok := node.postParseParagraphAsQuote(lines)
	if ok {
		return
	}
}

func (node *adocNode) postParseParagraphAsQuote(lines [][]byte) bool {
	lastLine := lines[len(lines)-1]
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
	firstLine := lines[0]
	if firstLine[0] != '"' {
		return false
	}
	secondLastLine := lines[len(lines)-2]
	if secondLastLine[len(secondLastLine)-1] != '"' {
		return false
	}

	node.raw = node.raw[:0]

	secondLastIdx := len(lines) - 2
	for x, line := range lines[:len(lines)-1] {
		if x == 0 {
			if x == secondLastIdx {
				node.Write(line[1 : len(line)-1])
			} else {
				node.Write(line[1:])
			}
		} else if x == secondLastIdx {
			node.Write(line[:len(line)-1])
		} else {
			node.Write(line)
		}
		node.WriteByte('\n')
	}

	node.kind = nodeKindBlockExcerpts
	opts := strings.SplitN(string(lastLine[3:]), `,`, 2)
	node.setQuoteOpts(opts)

	return true
}

func (node *adocNode) removeLastIfEmpty() {
	if node.child == nil {
		return
	}
	c := node
	for c.child != nil {
		c = c.child
		for c.next != nil {
			c = c.next
		}
	}
	if c.kind != nodeKindText || len(c.raw) > 0 {
		return
	}
	if c.prev != nil {
		c.prev.next = nil
		if c.prev.kind == nodeKindText {
			node.raw = bytes.TrimRight(node.raw, " \t")
		}
	} else if c.parent != nil {
		c.parent.child = nil
	}
	c.prev = nil
	c.parent = nil
}

func (node *adocNode) setQuoteOpts(opts []string) {
	if len(opts) >= 1 {
		node.key = strings.TrimSpace(opts[0])
	}
	if len(opts) >= 2 {
		node.value = strings.TrimSpace(opts[1])
	}
}

func (node *adocNode) setStyleAdmonition(admName string) {
	admName = strings.ToLower(admName)
	node.classes = append(node.classes, admName)
	node.rawLabel.WriteString(strings.Title(admName))
}

func (node *adocNode) toHTML(doc *Document, tmpl *template.Template, w io.Writer, isForToC bool) (err error) {
	switch node.kind {
	case lineKindAttribute:
		doc.attributes[node.key] = node.value

	case nodeKindCrossReference:
		href, ok := node.Attrs[attrNameHref]
		if !ok {
			title, ok := node.Attrs[attrNameTitle]
			if !ok {
				title, ok = node.Attrs[attrNameRefText]
			}
			href = doc.titleID[title]
		}
		_, err = fmt.Fprintf(w, _htmlCrossReference, href, node.raw)

	case nodeKindMacroTOC:
		if doc.tocIsEnabled && doc.tocPosition == metaValueMacro {
			err = doc.tocHTML(tmpl, w)
			if err != nil {
				return fmt.Errorf("toHTML: nodeKindMacroTOC: %w", err)
			}
		}
	case nodeKindPreamble:
		err = tmpl.ExecuteTemplate(w, "BEGIN_PREAMBLE", nil)
	case nodeKindSectionL1:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L1", node)
		if err != nil {
			return err
		}
		err = node.title.toHTML(doc, tmpl, w, isForToC)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("</h2>"))
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(`
<div class="sectionbody">`))

	case nodeKindSectionL2:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L2", node)
		if err != nil {
			return err
		}
		err = node.title.toHTML(doc, tmpl, w, isForToC)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("</h3>"))
	case nodeKindSectionL3:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L3", node)
		if err != nil {
			return err
		}
		if node.title != nil {
			err = node.title.toHTML(doc, tmpl, w, isForToC)
		}
		_, err = w.Write([]byte("</h4>"))
	case nodeKindSectionL4:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L4", node)
		if err != nil {
			return err
		}
		if node.title != nil {
			err = node.title.toHTML(doc, tmpl, w, isForToC)
		}
		_, err = w.Write([]byte("</h5>"))
	case nodeKindSectionL5:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L5", node)
		if err != nil {
			return err
		}
		if node.title != nil {
			err = node.title.toHTML(doc, tmpl, w, isForToC)
		}
		_, err = w.Write([]byte("</h6>"))
	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_ADMONITION", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "BEGIN_PARAGRAPH", node)
		}

	case nodeKindLiteralParagraph, nodeKindBlockLiteral, nodeKindBlockLiteralNamed:
		err = tmpl.ExecuteTemplate(w, "BLOCK_LITERAL", node)
	case nodeKindBlockListing, nodeKindBlockListingNamed:
		err = tmpl.ExecuteTemplate(w, "BLOCK_LISTING", node)

	case nodeKindInlineImage:
		err = tmpl.ExecuteTemplate(w, "INLINE_IMAGE", node)
	case nodeKindListOrdered:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_ORDERED", node)
	case nodeKindListUnordered:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_UNORDERED", node)
	case nodeKindListDescription:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_DESCRIPTION", node)

	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		_, err = w.Write([]byte("\n<li>"))

	case nodeKindListDescriptionItem:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_DESCRIPTION_ITEM", node)
	case lineKindHorizontalRule:
		err = tmpl.ExecuteTemplate(w, "HORIZONTAL_RULE", nil)
	case lineKindPageBreak:
		err = tmpl.ExecuteTemplate(w, "PAGE_BREAK", nil)
	case nodeKindBlockExample:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_ADMONITION", node)
			if err != nil {
				return err
			}
			err = tmpl.ExecuteTemplate(w, "BLOCK_TITLE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "BEGIN_EXAMPLE", node)
		}
	case nodeKindBlockImage:
		err = tmpl.ExecuteTemplate(w, "BLOCK_IMAGE", node)
	case nodeKindBlockOpen:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_ADMONITION", node)
			if err != nil {
				return err
			}
			err = tmpl.ExecuteTemplate(w, "BLOCK_TITLE", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "BEGIN_BLOCK_OPEN", node)
		}
	case nodeKindBlockPassthrough:
		_, err = w.Write([]byte("\n"))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindBlockExcerpts:
		if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "BEGIN_QUOTE", node)
		}
	case nodeKindBlockSidebar:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SIDEBAR", node)
	case nodeKindBlockVideo:
		err = tmpl.ExecuteTemplate(w, "BLOCK_VIDEO", node)
	case nodeKindBlockAudio:
		err = tmpl.ExecuteTemplate(w, "BLOCK_AUDIO", node)

	case nodeKindInlineID:
		if !isForToC {
			err = tmpl.ExecuteTemplate(w, "INLINE_ID", node)
		}

	case nodeKindInlineIDShort:
		if !isForToC {
			err = tmpl.ExecuteTemplate(w, "BEGIN_INLINE_ID_SHORT", node)
			if err != nil {
				return err
			}
			_, err = w.Write(node.raw)
		}

	case nodeKindInlineParagraph:
		_, err = w.Write([]byte("\n<p>"))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindPassthrough:
		_, err = w.Write(node.raw)
	case nodeKindPassthroughDouble:
		_, err = w.Write(node.raw)
	case nodeKindPassthroughTriple:
		_, err = w.Write(node.raw)

	case nodeKindSymbolQuoteDoubleBegin:
		_, err = w.Write([]byte(symbolQuoteDoubleBegin))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindSymbolQuoteDoubleEnd:
		_, err = w.Write([]byte(symbolQuoteDoubleEnd))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindSymbolQuoteSingleBegin:
		_, err = w.Write([]byte(symbolQuoteSingleBegin))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindSymbolQuoteSingleEnd:
		_, err = w.Write([]byte(symbolQuoteSingleEnd))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindText:
		_, err = w.Write(node.raw)
	case nodeKindTextBold:
		if node.HasStyle(styleTextBold) {
			_, err = w.Write([]byte("<strong>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("*"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindUnconstrainedBold:
		if node.HasStyle(styleTextBold) {
			_, err = w.Write([]byte("<strong>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("**"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindTextItalic:
		if node.HasStyle(styleTextItalic) {
			_, err = w.Write([]byte("<em>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("_"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindUnconstrainedItalic:
		if node.HasStyle(styleTextItalic) {
			_, err = w.Write([]byte("<em>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("__"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindTextMono:
		if node.HasStyle(styleTextMono) {
			_, err = w.Write([]byte("<code>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("`"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindUnconstrainedMono:
		if node.HasStyle(styleTextMono) {
			_, err = w.Write([]byte("<code>"))
		} else if len(node.raw) > 0 {
			_, err = w.Write([]byte("``"))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindURL:
		err = tmpl.ExecuteTemplate(w, "BEGIN_URL", node)
	case nodeKindTextSubscript:
		_, err = fmt.Fprintf(w, "<sub>%s</sub>", node.raw)
	case nodeKindTextSuperscript:
		_, err = fmt.Fprintf(w, "<sup>%s</sup>", node.raw)
	}
	if err != nil {
		return err
	}

	if node.child != nil {
		err = node.child.toHTML(doc, tmpl, w, isForToC)
		if err != nil {
			return err
		}
	}

	switch node.kind {
	case nodeKindPreamble:
		_, err = w.Write([]byte("\n</div>"))
		if err != nil {
			return fmt.Errorf("toHTML: nodeKindPreamble: %w", err)
		}
		if doc.tocIsEnabled && doc.tocPosition == metaValuePreamble {
			err = doc.tocHTML(tmpl, w)
			if err != nil {
				return fmt.Errorf("ToHTML: %w", err)
			}
		}
		_, err = w.Write([]byte("\n</div>"))

	case nodeKindSectionL1:
		_, err = w.Write([]byte("\n</div>\n</div>"))

	case nodeKindSectionL2, nodeKindSectionL3, nodeKindSectionL4, nodeKindSectionL5:
		_, err = w.Write([]byte("\n</div>"))

	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "END_ADMONITION", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "END_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "END_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "END_PARAGRAPH", node)
		}

	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		_, err = w.Write([]byte("\n</li>"))

	case nodeKindListDescriptionItem:
		err = tmpl.ExecuteTemplate(w, "END_LIST_DESCRIPTION_ITEM", node)
	case nodeKindListOrdered:
		err = tmpl.ExecuteTemplate(w, "END_LIST_ORDERED", nil)
	case nodeKindListUnordered:
		err = tmpl.ExecuteTemplate(w, "END_LIST_UNORDERED", nil)
	case nodeKindListDescription:
		err = tmpl.ExecuteTemplate(w, "END_LIST_DESCRIPTION", node)
	case nodeKindBlockExample:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "END_ADMONITION", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "END_EXAMPLE", node)
		}
	case nodeKindBlockOpen:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "END_ADMONITION", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "END_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "END_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "END_BLOCK_OPEN", node)
		}
	case nodeKindBlockExcerpts:
		if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "END_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "END_QUOTE", node)
		}
	case nodeKindBlockSidebar:
		err = tmpl.ExecuteTemplate(w, "END_SIDEBAR", node)

	case nodeKindInlineIDShort:
		if !isForToC {
			err = tmpl.ExecuteTemplate(w, "END_INLINE_ID_SHORT", node)
		}

	case nodeKindInlineParagraph:
		_, err = w.Write([]byte("</p>"))

	case nodeKindTextBold, nodeKindUnconstrainedBold:
		if node.HasStyle(styleTextBold) {
			_, err = fmt.Fprintf(w, "</strong>")
		}
	case nodeKindTextItalic, nodeKindUnconstrainedItalic:
		if node.HasStyle(styleTextItalic) {
			_, err = fmt.Fprintf(w, "</em>")
		}
	case nodeKindTextMono, nodeKindUnconstrainedMono:
		if node.HasStyle(styleTextMono) {
			_, err = fmt.Fprintf(w, "</code>")
		}
	case nodeKindURL:
		err = tmpl.ExecuteTemplate(w, "END_URL", node)
	}
	if err != nil {
		return err
	}

	if node.next != nil {
		err = node.next.toHTML(doc, tmpl, w, isForToC)
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *adocNode) toText(w io.Writer) (err error) {
	switch node.kind {
	case nodeKindPassthrough:
		_, err = w.Write(node.raw)
	case nodeKindPassthroughDouble:
		_, err = w.Write(node.raw)
	case nodeKindPassthroughTriple:
		_, err = w.Write(node.raw)

	case nodeKindSymbolQuoteDoubleBegin:
		_, err = w.Write([]byte(symbolQuoteDoubleBegin))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindSymbolQuoteDoubleEnd:
		_, err = w.Write([]byte(symbolQuoteDoubleEnd))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindSymbolQuoteSingleBegin:
		_, err = w.Write([]byte(symbolQuoteSingleBegin))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)
	case nodeKindSymbolQuoteSingleEnd:
		_, err = w.Write([]byte(symbolQuoteSingleEnd))
		if err != nil {
			return err
		}
		_, err = w.Write(node.raw)

	case nodeKindText:
		_, err = w.Write(node.raw)
	case nodeKindTextBold:
		if !node.HasStyle(styleTextBold) {
			_, err = w.Write([]byte("*"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)
	case nodeKindUnconstrainedBold:
		if !node.HasStyle(styleTextBold) {
			_, err = w.Write([]byte("**"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)

	case nodeKindTextItalic:
		if !node.HasStyle(styleTextItalic) {
			_, err = w.Write([]byte("_"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)
	case nodeKindUnconstrainedItalic:
		if !node.HasStyle(styleTextItalic) {
			_, err = w.Write([]byte("__"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)

	case nodeKindTextMono:
		if !node.HasStyle(styleTextMono) {
			_, err = w.Write([]byte("`"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)

	case nodeKindUnconstrainedMono:
		if !node.HasStyle(styleTextMono) {
			_, err = w.Write([]byte("``"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(node.raw)

	case nodeKindURL:
		_, err = w.Write(node.raw)
	case nodeKindTextSubscript:
		_, err = w.Write(node.raw)
	case nodeKindTextSuperscript:
		_, err = w.Write(node.raw)
	}
	if err != nil {
		return err
	}

	if node.child != nil {
		err = node.child.toText(w)
		if err != nil {
			return err
		}
	}
	if node.next != nil {
		err = node.next.toText(w)
		if err != nil {
			return err
		}
	}
	return nil
}
