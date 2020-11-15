// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	classNameArticle      = "article"
	classNameListingBlock = "listingblock"
	classNameLiteralBlock = "literalblock"
	classNameToc          = "toc"
	classNameToc2         = "toc2"
	classNameTocLeft      = "toc-left"
	classNameTocRight     = "toc-right"
	classNameUlist        = "ulist"
)

const (
	htmlSymbolAmpersand        = "&amp;"
	htmlSymbolApostrophe       = "&#8217;"
	htmlSymbolCopyright        = "&#169;"
	htmlSymbolDoubleLeftArrow  = "&#8656;"
	htmlSymbolDoubleRightArrow = "&#8658;"
	htmlSymbolEllipsis         = "&#8230;"
	htmlSymbolEmdash           = "&#8212;"
	htmlSymbolGreaterthan      = "&gt;"
	htmlSymbolLessthan         = "&lt;"
	htmlSymbolRegistered       = "&#174;"
	htmlSymbolSingleLeftArrow  = "&#8592;"
	htmlSymbolSingleRightArrow = "&#8594;"
	htmlSymbolTrademark        = "&#8482;"
)

func htmlWriteBlockBegin(node *adocNode, out io.Writer, addClass string) {
	fmt.Fprint(out, "\n<div")

	if len(node.ID) > 0 {
		fmt.Fprintf(out, ` id="%s"`, node.ID)
	}

	classes := node.Classes()
	c := strings.TrimSpace(addClass + " " + classes)
	if len(c) > 0 {
		fmt.Fprintf(out, ` class="%s">`, c)
	} else {
		fmt.Fprint(out, ">")
	}

	if !(node.IsStyleAdmonition() ||
		node.kind == nodeKindBlockImage ||
		node.kind == nodeKindBlockExample ||
		node.kind == nodeKindBlockSidebar) &&
		len(node.rawTitle) > 0 {

		fmt.Fprintf(out, "\n<div class=%q>%s</div>",
			attrValueTitle, node.rawTitle)
	}
}

func htmlWriteBlockAdmonition(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "admonitionblock")

	fmt.Fprint(out, "\n<table>\n<tr>\n<td class=\"icon\">")

	iconsFont := node.Attrs[attrNameIcons]
	if iconsFont == attrValueFont {
		fmt.Fprintf(out, _htmlAdmonitionIconsFont,
			strings.ToLower(node.Classes()), node.rawLabel.String())
	} else {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			node.rawLabel.String())
	}

	fmt.Fprintf(out, _htmlAdmonitionContent, node.raw)

	if len(node.rawTitle) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			node.rawTitle)
	}
}

func htmlWriteBlockAudio(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "audioblock")

	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)

	src := node.Attrs[attrNameSrc]

	optAutoplay, ok := node.Opts[optNameAutoplay]
	if ok {
		optAutoplay = " autoplay"
	}
	optControls, ok := node.Opts[optNameControls]
	if ok {
		optControls = " controls"
	}
	optLoop, ok := node.Opts[optNameLoop]
	if ok {
		optLoop = " loop"
	}

	fmt.Fprintf(out, _htmlBlockAudio, src, optAutoplay, optControls, optLoop)
}

func htmlWriteBlockExample(doc *Document, node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "exampleblock")
	if len(node.rawTitle) > 0 {
		doc.counterExample++
		fmt.Fprintf(out, "\n<div class=%q>Example %d. %s</div>",
			attrValueTitle, doc.counterExample, node.rawTitle)
	}
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
}

func htmlWriteBlockImage(doc *Document, node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "imageblock")

	src := node.Attrs[attrNameSrc]
	alt := node.Attrs[attrNameAlt]

	var width, height string
	v, ok := node.Attrs[attrNameWidth]
	if ok && len(v) > 0 {
		width = ` width="` + v + `"`
	}
	v, ok = node.Attrs[attrNameHeight]
	if ok && len(v) > 0 {
		height = ` height="` + v + `"`
	}

	fmt.Fprintf(out, _htmlBlockImage, src, alt, width, height)

	if len(node.rawTitle) > 0 {
		doc.counterImage++
		fmt.Fprintf(out, "\n<div class=%q>Figure %d. %s</div>",
			attrValueTitle, doc.counterImage, node.rawTitle)
	}

	fmt.Fprint(out, "\n</div>")
}

func htmlWriteBlockLiteral(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "")
	fmt.Fprintf(out, _htmlBlockLiteralContent, node.raw)
}

func htmlWriteBlockOpenBegin(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "openblock")
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
}

func htmlWriteBlockQuote(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "quoteblock")
	fmt.Fprintf(out, "\n<blockquote>\n%s", node.raw)
}

func htmlWriteBlockQuoteEnd(node *adocNode, out io.Writer) {
	fmt.Fprint(out, "\n</blockquote>")
	if len(node.key) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrValueAttribution, node.key)
	}
	if len(node.value) > 0 {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", node.value)
	}
	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBlockSidebar(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "sidebarblock")
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
	if len(node.rawTitle) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			node.rawTitle)
	}
}

func htmlWriteBlockVerse(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "verseblock")
	fmt.Fprintf(out, "\n<pre class=%q>%s", attrValueContent, node.raw)
}

func htmlWriteBlockVerseEnd(node *adocNode, out io.Writer) {
	fmt.Fprint(out, "</pre>")
	if len(node.key) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrValueAttribution, node.key)
	}
	if len(node.value) > 0 {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", node.value)
	}
	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBlockVideo(node *adocNode, out io.Writer) {
	src := node.GetVideoSource()
	width, withWidth := node.Attrs[attrNameWidth]
	if withWidth {
		width = fmt.Sprintf(` width="%s"`, width)
	}
	height, withHeight := node.Attrs[attrNameHeight]
	if withHeight {
		height = fmt.Sprintf(` height="%s"`, height)
	}
	_, isYoutube := node.Attrs[attrNameYoutube]
	_, isVimeo := node.Attrs[attrNameVimeo]

	htmlWriteBlockBegin(node, out, "videoblock")

	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)

	if isYoutube {
		optFullscreen, noFullscreen := node.Attrs[optVideoNofullscreen]
		if !noFullscreen {
			optFullscreen = " allowfullscreen"
		}
		fmt.Fprintf(out, _htmlBlockVideoYoutube, width, height, src, optFullscreen)
	} else if isVimeo {
		fmt.Fprintf(out, _htmlBlockVideoVimeo, width, height, src)
	} else {
		optPoster, withPoster := node.Attrs[attrNamePoster]
		if withPoster {
			optPoster = fmt.Sprintf(` poster="%s"`, optPoster)
		}
		optControls, ok := node.Attrs[optNameNocontrols]
		if !ok {
			optControls = " controls"
		}
		optAutoplay, ok := node.Attrs[optNameAutoplay]
		if ok {
			optAutoplay = " autoplay"
		}
		optLoop, ok := node.Attrs[optNameLoop]
		if ok {
			optLoop = " loop"
		}

		fmt.Fprintf(out, _htmlBlockVideo, src, width,
			height, optPoster, optControls, optAutoplay, optLoop)
	}

	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBody(doc *Document, out *bytes.Buffer) {
	fmt.Fprint(out, "\n<div id=\"content\">")

	if doc.content.child != nil {
		doc.content.child.toHTML(doc, out, false)
	}
	if doc.content.next != nil {
		doc.content.next.toHTML(doc, out, false)
	}

	fmt.Fprint(out, "\n</div>")
}

func htmlWriteFooter(doc *Document, out io.Writer) {
	fmt.Fprint(out, `
<div id="footer">
<div id="footer-text">`)

	if len(doc.Revision.Number) > 0 {
		fmt.Fprintf(out, "\nVersion %s<br>", doc.Revision.Number)
	}

	if len(doc.LastUpdated) > 0 {
		fmt.Fprintf(out, "\nLast updated %s", doc.LastUpdated)
	}

	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteHeader(doc *Document, out io.Writer) {
	fmt.Fprint(out, "\n<div id=\"header\">")

	_, ok := doc.Attributes[metaNameShowTitle]
	if ok {
		fmt.Fprint(out, "\n<h1>")
		doc.Title.node.toHTML(doc, out, false)
		fmt.Fprint(out, "</h1>")
	}

	fmt.Fprint(out, "\n<div class=\"details\">")

	var authorID, emailID string
	for x, author := range doc.Authors {
		if x == 0 {
			authorID = attrValueAuthor
			emailID = attrValueEmail
		} else {
			authorID = fmt.Sprintf("%s%d", attrValueAuthor, x+1)
			emailID = fmt.Sprintf("%s%d", attrValueEmail, x+1)
		}

		fmt.Fprintf(out, "\n<span id=%q class=%q>%s</span><br>",
			authorID, attrValueAuthor, author.FullName())

		if len(author.Email) > 0 {
			fmt.Fprintf(out, "\n<span id=%q class=%q><a href=\"mailto:%s\">%s</a></span><br>",
				emailID, attrValueEmail, author.Email,
				author.Email)
		}
	}

	if len(doc.Revision.Number) > 0 {
		prefix, ok := doc.Attributes[metaNameVersionLabel]
		if ok && len(prefix) == 0 {
			prefix = defVersionPrefix
		} else {
			prefix = " "
		}

		sep := ""
		if len(doc.Revision.Date) > 0 {
			sep = ","
		}

		fmt.Fprintf(out, "\n<span id=%q>%s%s%s</span>",
			attrValueRevNumber, prefix, doc.Revision.Number, sep)
	}
	if len(doc.Revision.Date) > 0 {
		fmt.Fprintf(out, "\n<span id=%q>%s</span>", attrValueRevDate,
			doc.Revision.Date)
	}
	if len(doc.Revision.Remark) > 0 {
		fmt.Fprintf(out, "\n<br><span id=%q>%s</span>",
			metaNameRevRemark, doc.Revision.Remark)
	}
	fmt.Fprint(out, "\n</div>")

	if doc.tocIsEnabled && (doc.tocPosition == "" ||
		doc.tocPosition == metaValueAuto ||
		doc.tocPosition == metaValueLeft ||
		doc.tocPosition == metaValueRight) {
		doc.tocHTML(out)
	}
	fmt.Fprint(out, "\n</div>")
}

func htmlWriteInlineImage(node *adocNode, out io.Writer) {
	classes := strings.TrimSpace("image " + node.Classes())
	fmt.Fprintf(out, "<span class=%q>", classes)
	link, withLink := node.Attrs[attrNameLink]
	if withLink {
		fmt.Fprintf(out, "<a class=%q href=%q>", attrValueImage, link)
	}

	src := node.Attrs[attrNameSrc]
	alt := node.Attrs[attrNameAlt]

	width, ok := node.Attrs[attrNameWidth]
	if ok {
		width = fmt.Sprintf(` width="%s"`, width)
	}
	height, ok := node.Attrs[attrNameHeight]
	if ok {
		height = fmt.Sprintf(` height="%s"`, height)
	}

	fmt.Fprintf(out, "<img src=%q alt=%q%s%s>", src, alt, width, height)

	if withLink {
		fmt.Fprint(out, `</a>`)
	}

	fmt.Fprint(out, `</span>`)
}

func htmlWriteListDescription(node *adocNode, out io.Writer) {
	var openTag string
	if node.IsStyleQandA() {
		htmlWriteBlockBegin(node, out, "qlist qanda")
		openTag = "\n<ol>"
	} else if node.IsStyleHorizontal() {
		htmlWriteBlockBegin(node, out, "hdlist")
		openTag = "\n<table>"
	} else {
		htmlWriteBlockBegin(node, out, "dlist")
		openTag = "\n<dl>"
	}

	fmt.Fprint(out, openTag)
}

func htmlWriteListDescriptionEnd(node *adocNode, out io.Writer) {
	if node.IsStyleQandA() {
		fmt.Fprintf(out, "\n</ol>\n</div>")
	} else if node.IsStyleHorizontal() {
		fmt.Fprintf(out, "\n</table>\n</div>")
	} else {
		fmt.Fprintf(out, "\n</dl>\n</div>")
	}
}

func htmlWriteListOrdered(node *adocNode, out io.Writer) {
	class := node.getListOrderedClass()
	tipe := node.getListOrderedType()
	if len(tipe) > 0 {
		tipe = ` type="` + tipe + `"`
	}

	htmlWriteBlockBegin(node, out, "olist "+class)

	fmt.Fprintf(out, "\n<ol class=\"%s\"%s>", class, tipe)
}

func htmlWriteListOrderedEnd(out io.Writer) {
	fmt.Fprint(out, "\n</ol>\n</div>")
}

func htmlWriteListUnordered(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "")
	fmt.Fprintf(out, "\n<ul>")
}

func htmlWriteListUnorderedEnd(out io.Writer) {
	fmt.Fprint(out, "\n</ul>\n</div>")
}

func htmlWriteParagraphBegin(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "paragraph")
	fmt.Fprint(out, "\n<p>")
}

func htmlWriteSection(doc *Document, node *adocNode, out io.Writer, isForToC bool) {
	var class, tag string
	switch node.kind {
	case nodeKindSectionL1:
		class = "sect1"
		tag = "h2"
	case nodeKindSectionL2:
		class = "sect2"
		tag = "h3"
	case nodeKindSectionL3:
		class = "sect3"
		tag = "h4"
	case nodeKindSectionL4:
		class = "sect4"
		tag = "h5"
	case nodeKindSectionL5:
		class = "sect5"
		tag = "h6"
	}

	fmt.Fprintf(out, "\n<div class=%q>\n<%s id=%q>", class, tag, node.ID)

	if node.sectnums != nil && node.level <= doc.sectLevel {
		fmt.Fprint(out, node.sectnums.String())
	}

	node.title.toHTML(doc, out, isForToC)

	fmt.Fprintf(out, "</%s>", tag)
	if node.kind == nodeKindSectionL1 {
		fmt.Fprint(out, "\n<div class=\"sectionbody\">")
	}
}

func htmlWriteToC(doc *Document, node *adocNode, out io.Writer, level int) {
	var sectClass string

	switch node.kind {
	case nodeKindSectionL1:
		sectClass = "sectlevel1"
	case nodeKindSectionL2:
		sectClass = "sectlevel2"
	case nodeKindSectionL3:
		sectClass = "sectlevel3"
	case nodeKindSectionL4:
		sectClass = "sectlevel4"
	case nodeKindSectionL5:
		sectClass = "sectlevel5"
	}
	if node.level > doc.TOCLevel {
		sectClass = ""
	}

	if len(sectClass) > 0 {
		if level < node.level {
			fmt.Fprintf(out, "\n<ul class=\"%s\">", sectClass)
		} else if level > node.level {
			n := level
			for n > node.level {
				fmt.Fprint(out, "\n</ul>")
				n--
			}
		}

		fmt.Fprintf(out, "\n<li><a href=\"#%s\">", node.ID)

		if node.sectnums != nil {
			fmt.Fprint(out, node.sectnums.String())
		}

		node.title.toHTML(doc, out, true)
		fmt.Fprint(out, "</a>")
	}

	if node.child != nil {
		htmlWriteToC(doc, node.child, out, node.level)
	}
	if len(sectClass) > 0 {
		fmt.Fprint(out, "</li>")
	}
	if node.next != nil {
		htmlWriteToC(doc, node.next, out, node.level)
	}

	if len(sectClass) > 0 && level < node.level {
		fmt.Fprint(out, "\n</ul>\n")
	}
}

func htmlWriteURLBegin(node *adocNode, out io.Writer) {
	fmt.Fprintf(out, "<a href=\"%s\"", node.Attrs[attrNameHref])
	classes := node.Classes()
	if len(classes) > 0 {
		fmt.Fprintf(out, ` class="%s"`, classes)
	}
	target := node.Attrs[attrNameTarget]
	if len(target) > 0 {
		fmt.Fprintf(out, ` target="%s"`, target)
	}
	rel := node.Attrs[attrNameRel]
	if len(rel) > 0 {
		fmt.Fprintf(out, ` rel="%s"`, rel)
	}
	fmt.Fprintf(out, `>%s`, node.raw)
}

func htmlWriteURLEnd(out io.Writer) {
	fmt.Fprint(out, "</a>")
}
