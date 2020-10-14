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
	"text/template"
	"unicode"
)

//
// adocNode is the building block of asciidoc document.
//
type adocNode struct {
	kind     int
	level    int          // The number of dot for ordered list, or star '*' for unordered list.
	raw      bytes.Buffer // unparsed content of node.
	rawLabel bytes.Buffer
	rawTitle string
	style    int64
	classes  []string
	Attrs    map[string]string
	Opts     map[string]string
	key      string
	value    string

	parent *adocNode
	child  *adocNode
	next   *adocNode
	prev   *adocNode
}

func (node *adocNode) Classes() string {
	if len(node.classes) == 0 {
		return ""
	}
	return " " + strings.Join(node.classes, " ")
}

func (node *adocNode) Content() string {
	return strings.TrimRight(node.raw.String(), "\n")
}

func (node *adocNode) GenerateID(str string) string {
	id := make([]rune, 0, len(str)+1)
	id = append(id, '_')
	for _, c := range strings.ToLower(str) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			id = append(id, c)
		} else {
			if id[len(id)-1] != '_' {
				id = append(id, '_')
			}
		}
	}
	return strings.TrimRight(string(id), "_")
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
	u.RawQuery = strings.Join(q, "&")
	u.Fragment = fragment

	return u.String()
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

func (node *adocNode) applySubstitutions() {
	if len(node.rawTitle) > 0 {
		node.rawTitle = htmlSubstituteSpecialChars(node.rawTitle)
	}

	content := node.raw.String()
	content = strings.TrimRight(content, "\n")
	content = htmlSubstituteSpecialChars(content)
	node.raw.Reset()

	switch node.kind {
	case nodeKindBlockExample, nodeKindBlockExcerpts, nodeKindParagraph,
		nodeKindBlockSidebar:
		node.raw.WriteString(content)
	default:
		node.raw.WriteString(content)
	}
}

func (node *adocNode) debug(n int) {
	for x := 0; x < n; x++ {
		fmt.Printf("\t")
	}
	fmt.Printf("node: %3d %s\n", node.kind, node.raw.String())
	if node.child != nil {
		node.child.debug(n + 1)
	}
	if node.next != nil {
		node.next.debug(n)
	}
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
	attrs := parseBlockAttribute(line[attrBegin : attrEnd+1])
	node.Attrs = make(map[string]string, len(attrs)+1)
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

func (node *adocNode) parseImage(line string) bool {
	attrBegin := strings.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	attrEnd := strings.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}
	name := strings.TrimRight(line[:attrBegin], " \t")
	node.raw.WriteString(name)

	attrs := strings.Split(line[attrBegin+1:attrEnd], ",")
	if node.Attrs == nil {
		node.Attrs = make(map[string]string)
	}
	for x, attr := range attrs {
		switch x {
		case 0:
			alt := strings.TrimSpace(attrs[0])
			if len(alt) == 0 {
				dot := strings.IndexByte(name, '.')
				if dot > 0 {
					alt = name[:dot]
				}
			}
			node.Attrs[attrNameAlt] = alt
		case 1:
			node.Attrs[attrNameWidth] = attrs[1]
		case 2:
			node.Attrs[attrNameHeight] = attrs[2]
		default:
			kv := strings.SplitN(attr, "=", 2)
			if len(kv) != 2 {
				continue
			}
			var (
				ok  bool
				val = strings.Trim(kv[1], `"`)
			)
			switch kv[0] {
			case "float", "align", "role":
				ok = true
				if val == "center" {
					val = "text-center"
				}
			}
			if ok {
				if len(val) > 0 {
					node.classes = append(node.classes, val)
				}
			}
		}
	}
	return true
}

func (node *adocNode) parseLineAdmonition(line string) {
	sep := strings.IndexByte(line, ':')
	class := strings.ToLower(line[:sep])
	node.classes = append(node.classes, class)
	node.rawLabel.WriteString(strings.Title(class))
	line = strings.TrimSpace(line[sep+1:])
	node.raw.WriteString(line)
	node.raw.WriteByte('\n')
}

func (node *adocNode) parseListDescription(line string) {
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
		node.raw.WriteString(line[x:])
	}
}

func (node *adocNode) parseListOrdered(line string) {
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
	node.raw.WriteString(line[x:])
	node.raw.WriteByte('\n')
}

func (node *adocNode) parseListUnordered(line string) {
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
	node.raw.WriteString(line[x:])
	node.raw.WriteByte('\n')
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
	attrs := parseBlockAttribute(line[attrBegin : attrEnd+1])

	if node.Attrs == nil {
		node.Attrs = make(map[string]string, len(attrs)+1)
	}
	node.Attrs[attrNameSrc] = videoSrc

	var key, val string
	for x, attr := range attrs {
		kv := strings.Split(attr, "=")

		key = strings.ToLower(kv[0])
		if len(kv) >= 2 {
			val = kv[1]
		} else {
			val = "1"
		}

		if x == 0 {
			if key == attrNameYoutube {
				node.Attrs[key] = val
				continue
			}
			if key == attrNameVimeo {
				node.Attrs[key] = val
				continue
			}
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

	raw := bytes.TrimRight(node.raw.Bytes(), " \t\n")

	lines := bytes.Split(raw, []byte{'\n'})
	if len(lines) <= 1 {
		return
	}

	ok := node.postParseParagraphAsQuote(lines)
	if ok {
		return
	}

	node.raw.Reset()
	node.raw.Write(raw)
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

	node.raw.Reset()
	secondLastIdx := len(lines) - 2
	for x, line := range lines[:len(lines)-1] {
		if x == 0 {
			if x == secondLastIdx {
				node.raw.Write(line[1 : len(line)-1])
			} else {
				node.raw.Write(line[1:])
			}
		} else if x == secondLastIdx {
			node.raw.Write(line[:len(line)-1])
		} else {
			node.raw.Write(line)
		}
		node.raw.WriteByte('\n')
	}

	node.kind = nodeKindBlockExcerpts
	opts := strings.SplitN(string(lastLine[3:]), `,`, 2)
	node.setQuoteOpts(opts)

	return true
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

func (node *adocNode) toHTML(doc *Document, tmpl *template.Template, w io.Writer) (err error) {
	switch node.kind {
	case lineKindAttribute:
		doc.attributes[node.key] = node.value
	case nodeKindPreamble:
		err = tmpl.ExecuteTemplate(w, "BEGIN_PREAMBLE", nil)
	case nodeKindSectionL1:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L1", node)
	case nodeKindSectionL2:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L2", node)
	case nodeKindSectionL3:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L3", node)
	case nodeKindSectionL4:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L4", node)
	case nodeKindSectionL5:
		err = tmpl.ExecuteTemplate(w, "BEGIN_SECTION_L5", node)
	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_ADMONITION", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "BEGIN_VERSE", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "PARAGRAPH", node)
		}
	case nodeKindLiteralParagraph, nodeKindBlockLiteralNamed:
		err = tmpl.ExecuteTemplate(w, "BLOCK_LITERAL", node)
	case nodeKindBlockLiteralDelimiter:
		if node.IsStyleListing() {
			err = tmpl.ExecuteTemplate(w, "BLOCK_LISTING", node)
		} else {
			err = tmpl.ExecuteTemplate(w, "BLOCK_LITERAL", node)
		}
	case nodeKindBlockListingDelimiter:
		err = tmpl.ExecuteTemplate(w, "BLOCK_LISTING", node)
	case nodeKindListOrdered:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_ORDERED", node)
	case nodeKindListUnordered:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_UNORDERED", node)
	case nodeKindListDescription:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_DESCRIPTION", node)
	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		err = tmpl.ExecuteTemplate(w, "BEGIN_LIST_ITEM", node)
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
	}
	if err != nil {
		return err
	}

	if node.child != nil {
		err = node.child.toHTML(doc, tmpl, w)
		if err != nil {
			return err
		}
	}

	switch node.kind {
	case nodeKindPreamble:
		err = tmpl.ExecuteTemplate(w, "END_PREAMBLE", nil)
	case nodeKindSectionL1:
		err = tmpl.ExecuteTemplate(w, "END_SECTION_L1", nil)
	case nodeKindSectionL2, nodeKindSectionL3, nodeKindSectionL4, nodeKindSectionL5:
		err = tmpl.ExecuteTemplate(w, "END_SECTION", nil)
	case nodeKindParagraph:
		if node.IsStyleAdmonition() {
			err = tmpl.ExecuteTemplate(w, "END_ADMONITION", node)
		} else if node.IsStyleQuote() {
			err = tmpl.ExecuteTemplate(w, "END_QUOTE", node)
		} else if node.IsStyleVerse() {
			err = tmpl.ExecuteTemplate(w, "END_VERSE", node)
		}
	case nodeKindListOrderedItem, nodeKindListUnorderedItem:
		err = tmpl.ExecuteTemplate(w, "END_LIST_ITEM", nil)
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
	}
	if err != nil {
		return err
	}

	if node.next != nil {
		err = node.next.toHTML(doc, tmpl, w)
		if err != nil {
			return err
		}
	}

	return nil
}
