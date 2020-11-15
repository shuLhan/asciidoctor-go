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

		fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
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
		fmt.Fprintf(out, _htmlBlockTitle, node.rawLabel.String())
	}

	fmt.Fprintf(out, _htmlAdmonitionContent, node.raw)

	if len(node.rawTitle) > 0 {
		fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
	}
}

func htmlWriteBlockAudio(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "audioblock")

	fmt.Fprint(out, _htmlBlockContent)

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
		fmt.Fprintf(out, _htmlBlockExampleTitle, doc.counterExample, node.rawTitle)
	}
	fmt.Fprint(out, _htmlBlockContent)
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
		fmt.Fprintf(out, _htmlBlockImageTitle, doc.counterImage, node.rawTitle)
	}

	fmt.Fprint(out, _htmlBlockImageEnd)
}

func htmlWriteBlockLiteral(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "")
	fmt.Fprintf(out, _htmlBlockLiteralContent, node.raw)
}

func htmlWriteBlockOpenBegin(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "openblock")
	_, _ = fmt.Fprint(out, _htmlBlockContent)
}

func htmlWriteBlockQuote(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "quoteblock")
	_, _ = fmt.Fprintf(out, _htmlBlockQuoteBegin, node.raw)
}

func htmlWriteBlockQuoteEnd(node *adocNode, out io.Writer) {
	_, _ = fmt.Fprint(out, _htmlBlockQuoteEnd)
	if len(node.key) > 0 {
		_, _ = fmt.Fprintf(out, _htmlQuoteAuthor, node.key)
	}
	if len(node.value) > 0 {
		_, _ = fmt.Fprintf(out, _htmlQuoteCitation, node.value)
	}
	_, _ = fmt.Fprint(out, _htmlBlockEnd)
}

func htmlWriteBlockSidebar(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "sidebarblock")
	_, _ = fmt.Fprint(out, _htmlBlockContent)
	if len(node.rawTitle) > 0 {
		_, _ = fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
	}
}

func htmlWriteBlockVerse(node *adocNode, out io.Writer) {
	htmlWriteBlockBegin(node, out, "verseblock")
	_, _ = fmt.Fprintf(out, _htmlBlockVerse, node.raw)
}

func htmlWriteBlockVerseEnd(node *adocNode, out io.Writer) {
	_, _ = fmt.Fprint(out, _htmlBlockVerseEnd)
	if len(node.key) > 0 {
		_, _ = fmt.Fprintf(out, _htmlQuoteAuthor, node.key)
	}
	if len(node.value) > 0 {
		_, _ = fmt.Fprintf(out, _htmlQuoteCitation, node.value)
	}
	_, _ = fmt.Fprint(out, _htmlBlockEnd)
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

	fmt.Fprint(out, _htmlBlockContent)

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

	fmt.Fprint(out, _htmlBlockEnd)
}

func htmlWriteBody(doc *Document, out *bytes.Buffer) {
	htmlWriteHeader(doc, out)

	fmt.Fprint(out, _htmlContentBegin)

	if doc.content.child != nil {
		doc.content.child.toHTML(doc, out, false)
	}
	if doc.content.next != nil {
		doc.content.next.toHTML(doc, out, false)
	}

	fmt.Fprint(out, _htmlContentEnd)
}

func htmlWriteHeader(doc *Document, out io.Writer) {
	fmt.Fprint(out, _htmlHeaderBegin)

	fmt.Fprint(out, _htmlHeaderTitleBegin)
	doc.Title.node.toHTML(doc, out, false)
	fmt.Fprint(out, _htmlHeaderTitleEnd)

	fmt.Fprint(out, _htmlHeaderDetail)
	if len(doc.Author) > 0 {
		fmt.Fprintf(out, _htmlHeaderDetailAuthor, doc.Author)
	}
	if len(doc.RevNumber) > 0 {
		fmt.Fprintf(out, _htmlHeaderDetailRevNumber,
			doc.RevNumber, doc.RevSeparator)
	}
	if len(doc.RevDate) > 0 {
		fmt.Fprintf(out, _htmlHeaderDetailRevDate, doc.RevDate)
	}
	fmt.Fprint(out, _htmlHeaderDetailEnd)

	if doc.tocIsEnabled && (doc.tocPosition == "" ||
		doc.tocPosition == metaValueAuto ||
		doc.tocPosition == metaValueLeft ||
		doc.tocPosition == metaValueRight) {
		doc.tocHTML(out)
	}
	fmt.Fprint(out, _htmlHeaderEnd)
}

func htmlWriteInlineImage(node *adocNode, out io.Writer) {
	classes := strings.TrimSpace("image " + node.Classes())
	fmt.Fprintf(out, _htmlInlineImage, classes)
	link, withLink := node.Attrs[attrNameLink]
	if withLink {
		fmt.Fprintf(out, _htmlInlineImageLink, link)
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

	fmt.Fprintf(out, _htmlInlineImageImage, src, alt, width, height)

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

	fmt.Fprintf(out, _htmlSection, class, tag, node.ID)

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
