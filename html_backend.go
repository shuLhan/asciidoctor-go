// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	libstrings "github.com/shuLhan/share/lib/strings"
)

const (
	classNameArticle      = "article"
	classNameHalignCenter = "halign-center"
	classNameHalignLeft   = "halign-left"
	classNameHalignRight  = "halign-right"
	classNameListingBlock = "listingblock"
	classNameLiteral      = "literal"
	classNameLiteralBlock = "literalblock"
	classNameTableBlock   = "tableblock"
	classNameToc          = "toc"
	classNameToc2         = "toc2"
	classNameTocLeft      = "toc-left"
	classNameTocRight     = "toc-right"
	classNameUlist        = "ulist"
	classNameValignBottom = "valign-bottom"
	classNameValignMiddle = "valign-middle"
	classNameValignTop    = "valign-top"
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

	classes := node.htmlClasses()
	c := strings.TrimSpace(addClass + " " + classes)
	if len(c) > 0 {
		fmt.Fprintf(out, ` class="%s">`, c)
	} else {
		fmt.Fprint(out, ">")
	}

	if !(node.isStyleAdmonition() ||
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
			strings.ToLower(node.htmlClasses()), node.rawLabel.String())
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
	var (
		optAutoplay string
		optControls = " controls"
		optLoop     string
	)

	htmlWriteBlockBegin(node, out, "audioblock")

	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)

	src := node.Attrs[attrNameSrc]

	if libstrings.IsContain(node.options, optNameAutoplay) {
		optAutoplay = " autoplay"
	}
	if libstrings.IsContain(node.options, optNameNocontrols) {
		optControls = ""
	}
	if libstrings.IsContain(node.options, optNameLoop) {
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
	source, ok := node.Attrs[attrNameSource]
	if ok {
		class := "language-" + source
		fmt.Fprint(out, "\n<pre class=\"highlight\">")
		fmt.Fprintf(out, `<code class=%q data-lang=%q>%s</code></pre>`,
			class, source, node.raw)
		fmt.Fprint(out, "\n</div>\n</div>")
	} else {
		fmt.Fprintf(out, _htmlBlockLiteralContent, node.raw)
	}
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
	if v, ok := node.Attrs[attrNameAttribution]; ok {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrNameAttribution, v)
	}
	if v, ok := node.Attrs[attrNameCitation]; ok {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", v)
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
	if v, ok := node.Attrs[attrNameAttribution]; ok {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrNameAttribution, v)
	}
	if v, ok := node.Attrs[attrNameCitation]; ok {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", v)
	}
	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBlockVideo(node *adocNode, out io.Writer) {
	var (
		isYoutube bool
		isVimeo   bool
	)

	src := node.getVideoSource()
	width, withWidth := node.Attrs[attrNameWidth]
	if withWidth {
		width = fmt.Sprintf(` width="%s"`, width)
	}
	height, withHeight := node.Attrs[attrNameHeight]
	if withHeight {
		height = fmt.Sprintf(` height="%s"`, height)
	}

	if node.rawStyle == attrNameYoutube {
		isYoutube = true
	}
	if node.rawStyle == attrNameVimeo {
		isVimeo = true
	}

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
		var (
			optControls = " controls"
			optAutoplay string
			optLoop     string
		)

		optPoster, withPoster := node.Attrs[attrNamePoster]
		if withPoster {
			optPoster = fmt.Sprintf(` poster="%s"`, optPoster)
		}

		if libstrings.IsContain(node.options, optNameNocontrols) {
			optControls = ""
		}
		if libstrings.IsContain(node.options, optNameAutoplay) {
			optAutoplay = " autoplay"
		}
		if libstrings.IsContain(node.options, optNameLoop) {
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
		prefix, ok := doc.Attributes[metaNameVersionLabel]
		if ok && len(prefix) == 0 {
			prefix = "Version "
		} else {
			prefix = " "
		}

		fmt.Fprintf(out, "\n%s%s<br>", prefix, doc.Revision.Number)
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
		_, ok = doc.Attributes[metaNameNoTitle]
		if !ok {
			fmt.Fprint(out, "\n<h1>")
			doc.Title.node.toHTML(doc, out, false)
			fmt.Fprint(out, "</h1>")
		}
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
	classes := strings.TrimSpace("image " + node.htmlClasses())
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
	if node.isStyleQandA() {
		htmlWriteBlockBegin(node, out, "qlist qanda")
		openTag = "\n<ol>"
	} else if node.isStyleHorizontal() {
		htmlWriteBlockBegin(node, out, "hdlist")
		openTag = "\n<table>"
	} else {
		htmlWriteBlockBegin(node, out, "dlist")
		openTag = "\n<dl>"
	}

	fmt.Fprint(out, openTag)
}

func htmlWriteListDescriptionEnd(node *adocNode, out io.Writer) {
	if node.isStyleQandA() {
		fmt.Fprintf(out, "\n</ol>\n</div>")
	} else if node.isStyleHorizontal() {
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
	var classes string
	if len(node.rawStyle) != 0 {
		classes = fmt.Sprintf(" class=%q", node.rawStyle)
	}
	htmlWriteBlockBegin(node, out, "")
	fmt.Fprintf(out, "\n<ul%s>", classes)
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

	_, withSectAnchors := doc.Attributes[metaNameSectAnchors]
	if withSectAnchors {
		fmt.Fprintf(out, `<a class="anchor" href="#%s"></a>`, node.ID)
	}
	_, withSectlinks := doc.Attributes[metaNameSectLinks]
	if withSectlinks {
		fmt.Fprintf(out, `<a class="link" href="#%s">`, node.ID)
	}

	if node.sectnums != nil && node.level <= doc.sectLevel {
		fmt.Fprint(out, node.sectnums.String())
	}

	node.title.toHTML(doc, out, isForToC)

	if withSectlinks {
		fmt.Fprint(out, "</a>")
	}
	fmt.Fprintf(out, "</%s>", tag)

	if node.kind == nodeKindSectionL1 {
		fmt.Fprint(out, "\n<div class=\"sectionbody\">")
	}
}

func hmltWriteSectionDiscrete(doc *Document, node *adocNode, out io.Writer) {
	var (
		tag string
	)
	switch node.level {
	case nodeKindSectionL1:
		tag = "h2"
	case nodeKindSectionL2:
		tag = "h3"
	case nodeKindSectionL3:
		tag = "h4"
	case nodeKindSectionL4:
		tag = "h5"
	case nodeKindSectionL5:
		tag = "h6"
	}

	fmt.Fprintf(out, "\n<%s id=%q class=%q>", tag, node.ID, attrNameDiscrete)
	node.title.toHTML(doc, out, false)
	fmt.Fprintf(out, "</%s>", tag)
}

func htmlWriteTable(doc *Document, node *adocNode, out io.Writer) {
	var (
		footer *tableRow
		table  = node.table
	)

	if table == nil {
		return
	}

	fmt.Fprintf(out, "\n<table class=%q", table.classes.String())

	style := table.htmlStyle()
	if len(style) > 0 {
		fmt.Fprintf(out, " style=%q", style)
	}
	fmt.Fprint(out, ">")

	if len(node.rawTitle) > 0 {
		var (
			caption string
			ok      bool
		)

		doc.counterTable++
		_, withTableCaption := doc.Attributes[metaNameTableCaption]

		if withTableCaption {
			caption, ok = node.Attrs[attrNameCaption]
			if !ok {
				caption = fmt.Sprintf("Table %d.", doc.counterTable)
			}
		}
		fmt.Fprintf(out, "\n<caption class=%q>%s %s</caption>",
			attrValueTitle, caption, node.rawTitle)
	}

	fmt.Fprint(out, "\n<colgroup>")
	for _, format := range table.formats {
		if format.width != nil {
			fmt.Fprintf(out, "\n<col style=\"width: %s%%;\">",
				format.width)
		} else {
			fmt.Fprint(out, "\n<col>")
		}
	}
	fmt.Fprint(out, "\n</colgroup>")

	rows := table.rows
	if table.hasHeader {
		htmlWriteTableHeader(doc, rows[0], out)
		rows = rows[1:]
	}
	if table.hasFooter && len(rows) > 0 {
		footer = rows[len(rows)-1]
		rows = rows[:len(rows)-1]
	}

	if len(rows) > 0 {
		fmt.Fprint(out, "\n<tbody>")
		for _, row := range rows {
			htmlWriteTableRow(doc, table, row, out)
		}
		fmt.Fprint(out, "\n</tbody>")
	}

	if table.hasFooter && footer != nil {
		htmlWriteTableFooter(doc, table, footer, out)
	}

	fmt.Fprint(out, "\n</table>")
}

func htmlWriteTableFooter(doc *Document, table *adocTable, footer *tableRow, out io.Writer) {
	fmt.Fprint(out, "\n<tfoot>")
	htmlWriteTableRow(doc, table, footer, out)
	fmt.Fprint(out, "\n</tfoot>")

}

func htmlWriteTableHeader(doc *Document, header *tableRow, out io.Writer) {
	classRow := "tableblock halign-left valign-top"

	fmt.Fprint(out, "\n<thead>\n<tr>")
	for _, cell := range header.cells {
		fmt.Fprintf(out, "\n<th class=%q>", classRow)
		cont := parseInlineMarkup(doc, bytes.TrimSpace(cell.content))
		cont.toHTML(doc, out, false)
		fmt.Fprint(out, "</th>")
	}
	fmt.Fprint(out, "\n</tr>\n</thead>")
}

func htmlWriteTableRow(doc *Document, table *adocTable, row *tableRow, out io.Writer) {
	fmt.Fprint(out, "\n<tr>")
	for x, cell := range row.cells {
		format := table.formats[x]
		tag := "td"
		colspan := ""

		if format.style == colStyleHeader {
			tag = "th"
		}
		if cell.format.nspanCol > 0 {
			colspan = fmt.Sprintf(` colspan="%d"`, cell.format.nspanCol)
		}

		fmt.Fprintf(out, "\n<%s class=%q%s>", tag,
			format.htmlClasses(), colspan)

		contentTrimmed := bytes.TrimSpace(cell.content)

		switch format.style {
		case colStyleAsciidoc:
			subdoc := parseSub(doc, contentTrimmed)
			_ = subdoc.ToEmbeddedHTML(out)

		case colStyleDefault:
			rawParagraphs := bytes.Split(contentTrimmed, []byte("\n\n"))
			for x, p := range rawParagraphs {
				if x > 0 {
					fmt.Fprint(out, "\n")
				}
				fmt.Fprintf(out, "<p class=%q>", classNameTableBlock)
				container := parseInlineMarkup(doc, p)
				container.toHTML(doc, out, false)
				fmt.Fprint(out, "</p>")
			}

		case colStyleHeader, colStyleVerse:
			fmt.Fprintf(out, "<p class=%q>%s</p>",
				classNameTableBlock, contentTrimmed)

		case colStyleEmphasis:
			fmt.Fprintf(out, "<p class=%q><em>%s</em></p>",
				classNameTableBlock, contentTrimmed)

		case colStyleLiteral:
			fmt.Fprintf(out, "<div class=%q><pre>%s</pre></div>",
				classNameLiteral, cell.content)

		case colStyleMonospaced:
			fmt.Fprintf(out, "<p class=%q><code>%s</code></p>",
				classNameTableBlock, contentTrimmed)

		case colStyleStrong:
			fmt.Fprintf(out, "<p class=%q><strong>%s</strong></p>",
				classNameTableBlock, contentTrimmed)
		}

		fmt.Fprintf(out, "</%s>", tag)
	}
	fmt.Fprint(out, "\n</tr>")
}

func htmlWriteToC(doc *Document, node *adocNode, out io.Writer, level int) {
	var sectClass string

	isDiscrete := node.style&styleSectionDiscrete > 0

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

	if len(sectClass) > 0 && !isDiscrete {
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
	if len(sectClass) > 0 && !isDiscrete {
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
	classes := node.htmlClasses()
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
