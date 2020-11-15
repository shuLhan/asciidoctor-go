// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

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
	classes  attributeClass

	// The key and value for attribute (lineKindAttribute).
	key   string
	value string

	// title is the parsed rawTitle for section L1 or parsed raw for
	// section L2-L5.
	title *adocNode
	label *adocNode

	// sectnums contain the current section numbers.
	// It will be set only if attribute "sectnums" is on.
	sectnums *sectionCounters

	parent *adocNode
	child  *adocNode
	next   *adocNode
	prev   *adocNode
}

func (node *adocNode) Classes() string {
	if len(node.classes) == 0 {
		return ""
	}
	return node.classes.String()
}

func (node *adocNode) Content() string {
	node.raw = bytes.TrimRight(node.raw, "\n")
	return string(node.raw)
}

func (node *adocNode) getListOrderedClass() string {
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

func (node *adocNode) getListOrderedType() string {
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
				node.Attrs[optVideoNofullscreen] = ""
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
				node.Attrs[optNameNocontrols] = ""
				node.Attrs[opt] = ""
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

func (node *adocNode) parseBlockAudio(doc *Document, line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	src := strings.TrimRight(line[:attrBegin], " \t")
	_, _, attrs := parseAttributeElement(line[attrBegin : attrEnd+1])
	node.Attrs = make(map[string]string, len(attrs)+1)

	src = string(applySubstitutions(doc, []byte(src)))
	node.Attrs[attrNameSrc] = src

	var key, val string
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
// parseBlockImage parse the image block or line.
// The line parameter must not have "image::" block or "image:" macro prefix.
//
func (node *adocNode) parseBlockImage(doc *Document, line string) bool {
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
	src = string(applySubstitutions(doc, []byte(src)))
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
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(kv[1], `"`)
		switch key {
		case attrNameFloat, attrNameAlign, attrNameRole:
			if val == "center" {
				val = "text-center"
			}
			node.classes.add(val)
		default:
			node.Attrs[key] = val
		}
	}

	for k, v := range node.Attrs {
		if k == attrNameLink {
			v = string(applySubstitutions(doc, []byte(v)))
			node.Attrs[k] = v
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
	node.classes.add(class)
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
	node.Text = container.toText()

	if len(node.ID) == 0 {
		_, ok := doc.Attributes[metaNameSectIDs]
		if ok {
			node.ID = generateID(doc, node.Text)
			node.ID = doc.registerAnchor(node.ID, node.Text)
		}
	}

	refText, ok := node.Attrs[attrNameRefText]
	if ok {
		doc.titleID[refText] = node.ID
		// Replace the label with refText.
		anc := doc.anchors[node.ID]
		if anc != nil {
			anc.label = refText
		}
	}
	doc.titleID[node.Text] = node.ID

	_, ok = doc.Attributes[metaNameSectNums]
	if ok {
		node.sectnums = doc.sectnums.set(node.level)
	}
}

func (node *adocNode) parseStyleClass(line string) {
	line = strings.Trim(line, "[]")
	parts := strings.Split(line, ".")
	for _, class := range parts {
		class = strings.TrimSpace(class)
		if len(class) > 0 {
			node.classes.add(class)
		}
	}
}

func (node *adocNode) parseBlockVideo(doc *Document, line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	videoSrc := strings.TrimRight(line[:attrBegin], " \t")
	key, _, attrs := parseAttributeElement(line[attrBegin : attrEnd+1])

	if node.Attrs == nil {
		node.Attrs = make(map[string]string, len(attrs)+1)
	}
	videoSrc = string(applySubstitutions(doc, []byte(videoSrc)))
	node.Attrs[attrNameSrc] = videoSrc

	start := 0
	if key == attrNameYoutube || key == attrNameVimeo {
		node.Attrs[key] = ""
		start = 1
	}

	var val string
	for _, attr := range attrs[start:] {
		kv := strings.Split(attr, "=")

		key = strings.ToLower(kv[0])
		if len(kv) >= 2 {
			val = kv[1]
		} else {
			val = ""
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
			if item.kind == nodeKindListDescriptionItem {
				raw := item.rawLabel.Bytes()
				item.label = parseInlineMarkup(doc, raw)
				item.rawLabel.Reset()
			}
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

	node.postParseParagraphAsQuote(lines)
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
	node.classes.add(admName)
	node.rawLabel.WriteString(strings.Title(admName))
}

func (node *adocNode) toHTML(doc *Document, w io.Writer, isForToC bool) {
	switch node.kind {
	case lineKindAttribute:
		doc.Attributes.apply(node.key, node.value)

	case nodeKindCrossReference:
		href, ok := node.Attrs[attrNameHref]
		if !ok {
			title, ok := node.Attrs[attrNameTitle]
			if !ok {
				title = node.Attrs[attrNameRefText]
			}
			href = doc.titleID[title]
		}
		fmt.Fprintf(w, "<a href=\"#%s\">%s</a>", href, node.raw)

	case nodeKindMacroTOC:
		if doc.tocIsEnabled && doc.tocPosition == metaValueMacro {
			doc.tocHTML(w)
		}

	case nodeKindPreamble:
		fmt.Fprint(w, _htmlPreambleBegin)

	case nodeKindSectionL1, nodeKindSectionL2, nodeKindSectionL3,
		nodeKindSectionL4, nodeKindSectionL5:
		htmlWriteSection(doc, node, w, isForToC)

	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			htmlWriteBlockAdmonition(node, w)
		} else if node.IsStyleQuote() {
			htmlWriteBlockQuote(node, w)
		} else if node.IsStyleVerse() {
			htmlWriteBlockVerse(node, w)
		} else {
			htmlWriteParagraphBegin(node, w)
		}

	case nodeKindLiteralParagraph, nodeKindBlockLiteral,
		nodeKindBlockLiteralNamed,
		nodeKindBlockListing, nodeKindBlockListingNamed:
		htmlWriteBlockLiteral(node, w)

	case nodeKindInlineImage:
		htmlWriteInlineImage(node, w)

	case nodeKindListDescription:
		htmlWriteListDescription(node, w)
	case nodeKindListOrdered:
		htmlWriteListOrdered(node, w)
	case nodeKindListUnordered:
		htmlWriteListUnordered(node, w)

	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		fmt.Fprint(w, "\n<li>")

	case nodeKindListDescriptionItem:
		var (
			format string
			label  bytes.Buffer
		)
		if node.label != nil {
			node.label.toHTML(doc, &label, false)
		} else {
			label.Write(node.rawLabel.Bytes())
		}

		if node.IsStyleQandA() {
			format = _htmlListDescriptionItemQandABegin
		} else if node.IsStyleHorizontal() {
			format = _htmlListDescriptionItemHorizontalBegin
		} else {
			format = _htmlListDescriptionItemBegin
		}
		fmt.Fprintf(w, format, label.String())

	case lineKindHorizontalRule:
		fmt.Fprint(w, "\n<hr>")

	case lineKindPageBreak:
		fmt.Fprint(w, "\n<div style=\"page-break-after: always;\"></div>")

	case nodeKindBlockExample:
		if node.IsStyleAdmonition() {
			htmlWriteBlockAdmonition(node, w)
		} else {
			htmlWriteBlockExample(doc, node, w)
		}

	case nodeKindBlockImage:
		htmlWriteBlockImage(doc, node, w)

	case nodeKindBlockOpen:
		if node.IsStyleAdmonition() {
			htmlWriteBlockAdmonition(node, w)
		} else if node.IsStyleQuote() {
			htmlWriteBlockQuote(node, w)
		} else if node.IsStyleVerse() {
			htmlWriteBlockVerse(node, w)
		} else {
			htmlWriteBlockOpenBegin(node, w)
		}

	case nodeKindBlockPassthrough:
		fmt.Fprintf(w, "\n%s", node.raw)

	case nodeKindBlockExcerpts:
		if node.IsStyleVerse() {
			htmlWriteBlockVerse(node, w)
		} else {
			htmlWriteBlockQuote(node, w)
		}

	case nodeKindBlockSidebar:
		htmlWriteBlockSidebar(node, w)

	case nodeKindBlockVideo:
		htmlWriteBlockVideo(node, w)

	case nodeKindBlockAudio:
		htmlWriteBlockAudio(node, w)

	case nodeKindInlineID:
		if !isForToC {
			fmt.Fprintf(w, "<a id=%q></a>", node.ID)
		}

	case nodeKindInlineIDShort:
		if !isForToC {
			fmt.Fprintf(w, "<span id=%q>%s", node.ID, node.raw)
		}

	case nodeKindInlineParagraph:
		fmt.Fprintf(w, "\n<p>%s", node.raw)

	case nodeKindPassthrough:
		fmt.Fprint(w, string(node.raw))
	case nodeKindPassthroughDouble:
		fmt.Fprint(w, string(node.raw))
	case nodeKindPassthroughTriple:
		fmt.Fprint(w, string(node.raw))

	case nodeKindSymbolQuoteDoubleBegin:
		fmt.Fprint(w, symbolQuoteDoubleBegin, string(node.raw))
	case nodeKindSymbolQuoteDoubleEnd:
		fmt.Fprint(w, symbolQuoteDoubleEnd, string(node.raw))

	case nodeKindSymbolQuoteSingleBegin:
		fmt.Fprint(w, symbolQuoteSingleBegin, string(node.raw))
	case nodeKindSymbolQuoteSingleEnd:
		fmt.Fprint(w, symbolQuoteSingleEnd, string(node.raw))

	case nodeKindText:
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextBold:
		if node.HasStyle(styleTextBold) {
			fmt.Fprint(w, "<strong>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "*")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedBold:
		if node.HasStyle(styleTextBold) {
			fmt.Fprint(w, "<strong>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "**")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextItalic:
		if node.HasStyle(styleTextItalic) {
			fmt.Fprint(w, "<em>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "_")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedItalic:
		if node.HasStyle(styleTextItalic) {
			fmt.Fprint(w, "<em>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "__")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextMono:
		if node.HasStyle(styleTextMono) {
			fmt.Fprint(w, "<code>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "`")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedMono:
		if node.HasStyle(styleTextMono) {
			fmt.Fprint(w, "<code>")
		} else if len(node.raw) > 0 {
			fmt.Fprint(w, "``")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindURL:
		htmlWriteURLBegin(node, w)

	case nodeKindTextSubscript:
		fmt.Fprintf(w, "<sub>%s</sub>", node.raw)
	case nodeKindTextSuperscript:
		fmt.Fprintf(w, "<sup>%s</sup>", node.raw)
	}

	if node.child != nil {
		node.child.toHTML(doc, w, isForToC)
	}

	switch node.kind {
	case nodeKindPreamble:
		fmt.Fprint(w, "\n</div>")
		if doc.tocIsEnabled && doc.tocPosition == metaValuePreamble {
			doc.tocHTML(w)
		}
		fmt.Fprint(w, "\n</div>")

	case nodeKindSectionL1, nodeKindSectionL2, nodeKindSectionL3,
		nodeKindSectionL4, nodeKindSectionL5:
		if node.kind == nodeKindSectionL1 {
			fmt.Fprint(w, "\n</div>")
		}
		fmt.Fprint(w, "\n</div>")

	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if node.IsStyleQuote() {
			htmlWriteBlockQuoteEnd(node, w)
		} else if node.IsStyleVerse() {
			htmlWriteBlockVerseEnd(node, w)
		} else {
			fmt.Fprint(w, "</p>\n</div>")
		}

	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		fmt.Fprint(w, "\n</li>")

	case nodeKindListDescriptionItem:
		var format string
		if node.IsStyleQandA() {
			format = "\n</li>"
		} else if node.IsStyleHorizontal() {
			format = "\n</td>\n</tr>"
		} else {
			format = "\n</dd>"
		}
		fmt.Fprint(w, format)

	case nodeKindListDescription:
		htmlWriteListDescriptionEnd(node, w)
	case nodeKindListOrdered:
		htmlWriteListOrderedEnd(w)
	case nodeKindListUnordered:
		htmlWriteListUnorderedEnd(w)

	case nodeKindBlockExample:
		if node.IsStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else {
			fmt.Fprint(w, "\n</div>\n</div>")
		}

	case nodeKindBlockOpen:
		if node.IsStyleAdmonition() {
			fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if node.IsStyleQuote() {
			htmlWriteBlockQuoteEnd(node, w)
		} else if node.IsStyleVerse() {
			htmlWriteBlockVerseEnd(node, w)
		} else {
			fmt.Fprint(w, "\n</div>\n</div>")
		}
	case nodeKindBlockExcerpts:
		if node.IsStyleVerse() {
			htmlWriteBlockVerseEnd(node, w)
		} else {
			htmlWriteBlockQuoteEnd(node, w)
		}

	case nodeKindBlockSidebar:
		fmt.Fprint(w, "\n</div>\n</div>")

	case nodeKindInlineIDShort:
		if !isForToC {
			fmt.Fprint(w, "</span>")
		}

	case nodeKindInlineParagraph:
		fmt.Fprint(w, "</p>")

	case nodeKindTextBold, nodeKindUnconstrainedBold:
		if node.HasStyle(styleTextBold) {
			fmt.Fprint(w, "</strong>")
		}
	case nodeKindTextItalic, nodeKindUnconstrainedItalic:
		if node.HasStyle(styleTextItalic) {
			fmt.Fprint(w, "</em>")
		}
	case nodeKindTextMono, nodeKindUnconstrainedMono:
		if node.HasStyle(styleTextMono) {
			fmt.Fprint(w, "</code>")
		}
	case nodeKindURL:
		htmlWriteURLEnd(w)
	}

	if node.next != nil {
		node.next.toHTML(doc, w, isForToC)
	}
}

func (node *adocNode) toText() (text string) {
	var buf bytes.Buffer
	node.writeText(&buf)
	return buf.String()
}

func (node *adocNode) writeText(w io.Writer) {
	switch node.kind {
	case nodeKindPassthrough:
		fmt.Fprint(w, string(node.raw))
	case nodeKindPassthroughDouble:
		fmt.Fprint(w, string(node.raw))
	case nodeKindPassthroughTriple:
		fmt.Fprint(w, string(node.raw))

	case nodeKindSymbolQuoteDoubleBegin:
		fmt.Fprint(w, symbolQuoteDoubleBegin, string(node.raw))

	case nodeKindSymbolQuoteDoubleEnd:
		fmt.Fprint(w, symbolQuoteDoubleEnd, string(node.raw))

	case nodeKindSymbolQuoteSingleBegin:
		fmt.Fprint(w, symbolQuoteSingleBegin, string(node.raw))
	case nodeKindSymbolQuoteSingleEnd:
		fmt.Fprint(w, symbolQuoteSingleEnd, string(node.raw))

	case nodeKindText:
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextBold:
		if !node.HasStyle(styleTextBold) {
			fmt.Fprint(w, "*")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedBold:
		if !node.HasStyle(styleTextBold) {
			fmt.Fprint(w, "**")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextItalic:
		if !node.HasStyle(styleTextItalic) {
			fmt.Fprint(w, "_")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedItalic:
		if !node.HasStyle(styleTextItalic) {
			fmt.Fprint(w, "__")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindTextMono:
		if !node.HasStyle(styleTextMono) {
			fmt.Fprint(w, "`")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindUnconstrainedMono:
		if !node.HasStyle(styleTextMono) {
			fmt.Fprint(w, "``")
		}
		fmt.Fprint(w, string(node.raw))

	case nodeKindURL:
		fmt.Fprint(w, string(node.raw))
	case nodeKindTextSubscript:
		fmt.Fprint(w, string(node.raw))
	case nodeKindTextSuperscript:
		fmt.Fprint(w, string(node.raw))
	}

	if node.child != nil {
		node.child.writeText(w)
	}
	if node.next != nil {
		node.next.writeText(w)
	}
}
