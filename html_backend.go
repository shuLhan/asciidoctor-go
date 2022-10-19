// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	libstrings "github.com/shuLhan/share/lib/strings"
)

const (
	classNameArticle      = `article`
	classNameHalignCenter = `halign-center`
	classNameHalignLeft   = `halign-left`
	classNameHalignRight  = `halign-right`
	classNameListingBlock = `listingblock`
	classNameLiteral      = `literal`
	classNameLiteralBlock = `literalblock`
	classNameTableBlock   = `tableblock`
	classNameToc          = `toc`
	classNameToc2         = `toc2`
	classNameTocLeft      = `toc-left`
	classNameTocRight     = `toc-right`
	classNameUlist        = `ulist`
	classNameValignBottom = `valign-bottom`
	classNameValignMiddle = `valign-middle`
	classNameValignTop    = `valign-top`
)

const (
	htmlSymbolAmpersand         = `&amp;`
	htmlSymbolApostrophe        = `&#8217;`
	htmlSymbolBrokenVerticalBar = `&#166;`
	htmlSymbolCopyright         = `&#169;`
	htmlSymbolDegreeSign        = `&#176;`
	htmlSymbolDoubleLeftArrow   = `&#8656;`
	htmlSymbolDoubleQuote       = `&#34;`
	htmlSymbolDoubleRightArrow  = `&#8658;`
	htmlSymbolEllipsis          = `&#8230;`
	htmlSymbolEmdash            = `&#8212;`
	htmlSymbolGreaterthan       = `&gt;`
	htmlSymbolLeftDoubleQuote   = `&#8220;`
	htmlSymbolLeftSingleQuote   = `&#8216;`
	htmlSymbolLessthan          = `&lt;`
	htmlSymbolNonBreakingSpace  = `&#160;`
	htmlSymbolPlus              = `&#43;`
	htmlSymbolRegistered        = `&#174;`
	htmlSymbolRightDoubleQuote  = `&#8221;`
	htmlSymbolRightSingleQuote  = `&#8217;`
	htmlSymbolSingleLeftArrow   = `&#8592;`
	htmlSymbolSingleQuote       = `&#39;`
	htmlSymbolSingleRightArrow  = `&#8594;`
	htmlSymbolThinSpace         = `&#8201;`
	htmlSymbolTrademark         = `&#8482;`
	htmlSymbolWordJoiner        = `&#8288;`
	htmlSymbolZeroWidthSpace    = `&#8203;`
)

func htmlWriteBlockBegin(el *element, out io.Writer, addClass string) {
	fmt.Fprint(out, "\n<div")

	if len(el.ID) > 0 {
		fmt.Fprintf(out, ` id="%s"`, el.ID)
	}

	var (
		classes string = el.htmlClasses()
		c       string = strings.TrimSpace(addClass + ` ` + classes)
	)

	if len(c) > 0 {
		fmt.Fprintf(out, ` class="%s">`, c)
	} else {
		fmt.Fprint(out, `>`)
	}

	if !(el.isStyleAdmonition() ||
		el.kind == elKindBlockImage ||
		el.kind == elKindBlockExample ||
		el.kind == elKindBlockSidebar) &&
		len(el.rawTitle) > 0 {

		fmt.Fprintf(out, "\n<div class=%q>%s</div>",
			attrValueTitle, el.rawTitle)
	}
}

func htmlWriteBlockAdmonition(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `admonitionblock`)

	fmt.Fprint(out, "\n<table>\n<tr>\n<td class=\"icon\">")

	var (
		iconsFont string = el.Attrs[attrNameIcons]
	)

	if iconsFont == attrValueFont {
		fmt.Fprintf(out, _htmlAdmonitionIconsFont,
			strings.ToLower(el.htmlClasses()), el.rawLabel.String())
	} else {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			el.rawLabel.String())
	}

	fmt.Fprintf(out, _htmlAdmonitionContent, el.raw)

	if len(el.rawTitle) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			el.rawTitle)
	}
}

func htmlWriteBlockAudio(el *element, out io.Writer) {
	var (
		optControls string = ` controls`
		src         string = el.Attrs[attrNameSrc]

		optAutoplay string
		optLoop     string
	)

	htmlWriteBlockBegin(el, out, `audioblock`)

	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)

	if libstrings.IsContain(el.options, optNameAutoplay) {
		optAutoplay = ` autoplay`
	}
	if libstrings.IsContain(el.options, optNameNocontrols) {
		optControls = ``
	}
	if libstrings.IsContain(el.options, optNameLoop) {
		optLoop = ` loop`
	}

	fmt.Fprintf(out, _htmlBlockAudio, src, optAutoplay, optControls, optLoop)
}

func htmlWriteBlockExample(doc *Document, el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `exampleblock`)
	if len(el.rawTitle) > 0 {
		doc.counterExample++
		fmt.Fprintf(out, "\n<div class=%q>Example %d. %s</div>",
			attrValueTitle, doc.counterExample, el.rawTitle)
	}
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
}

func htmlWriteBlockImage(doc *Document, el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `imageblock`)

	var (
		src = el.Attrs[attrNameSrc]
		alt = el.Attrs[attrNameAlt]

		v      string
		width  string
		height string
		ok     bool
	)

	v, ok = el.Attrs[attrNameWidth]
	if ok && len(v) > 0 {
		width = ` width="` + v + `"`
	}
	v, ok = el.Attrs[attrNameHeight]
	if ok && len(v) > 0 {
		height = ` height="` + v + `"`
	}

	fmt.Fprintf(out, _htmlBlockImage, src, alt, width, height)

	if len(el.rawTitle) > 0 {
		doc.counterImage++
		fmt.Fprintf(out, "\n<div class=%q>Figure %d. %s</div>",
			attrValueTitle, doc.counterImage, el.rawTitle)
	}

	fmt.Fprint(out, "\n</div>")
}

func htmlWriteBlockLiteral(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, ``)

	var (
		source string
		class  string
		ok     bool
	)
	source, ok = el.Attrs[attrNameSource]
	if ok {
		class = `language-` + source
		fmt.Fprint(out, "\n<div class=\"content\">\n<pre class=\"highlight\">")
		fmt.Fprintf(out, `<code class=%q data-lang=%q>%s</code></pre>`,
			class, source, el.raw)
		fmt.Fprint(out, "\n</div>\n</div>")
	} else {
		fmt.Fprintf(out, _htmlBlockLiteralContent, el.raw)
	}
}

func htmlWriteBlockOpenBegin(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, "openblock")
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
}

func htmlWriteBlockQuote(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `quoteblock`)
	fmt.Fprintf(out, "\n<blockquote>\n%s", el.raw)
}

func htmlWriteBlockQuoteEnd(el *element, out io.Writer) {
	fmt.Fprint(out, "\n</blockquote>")

	var (
		v  string
		ok bool
	)

	v, ok = el.Attrs[attrNameAttribution]
	if ok {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrNameAttribution, v)
	}

	v, ok = el.Attrs[attrNameCitation]
	if ok {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", v)
	}
	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBlockSidebar(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `sidebarblock`)
	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)
	if len(el.rawTitle) > 0 {
		fmt.Fprintf(out, "\n<div class=%q>%s</div>", attrValueTitle,
			el.rawTitle)
	}
}

func htmlWriteBlockVerse(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, `verseblock`)
	fmt.Fprintf(out, "\n<pre class=%q>%s", attrValueContent, el.raw)
}

func htmlWriteBlockVerseEnd(el *element, out io.Writer) {
	fmt.Fprint(out, `</pre>`)

	var (
		v  string
		ok bool
	)

	v, ok = el.Attrs[attrNameAttribution]
	if ok {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s",
			attrNameAttribution, v)
	}

	v, ok = el.Attrs[attrNameCitation]
	if ok {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", v)
	}
	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBlockVideo(el *element, out io.Writer) {
	var (
		src    string
		width  string
		height string

		isYoutube  bool
		isVimeo    bool
		withWidth  bool
		withHeight bool
	)

	src = el.getVideoSource()
	width, withWidth = el.Attrs[attrNameWidth]
	if withWidth {
		width = fmt.Sprintf(` width="%s"`, width)
	}

	height, withHeight = el.Attrs[attrNameHeight]
	if withHeight {
		height = fmt.Sprintf(` height="%s"`, height)
	}

	if el.rawStyle == attrNameYoutube {
		isYoutube = true
	}
	if el.rawStyle == attrNameVimeo {
		isVimeo = true
	}

	htmlWriteBlockBegin(el, out, `videoblock`)

	fmt.Fprintf(out, "\n<div class=%q>", attrValueContent)

	if isYoutube {
		var (
			optFullscreen string
			noFullscreen  bool
		)

		optFullscreen, noFullscreen = el.Attrs[optVideoNofullscreen]
		if !noFullscreen {
			optFullscreen = ` allowfullscreen`
		}
		fmt.Fprintf(out, _htmlBlockVideoYoutube, width, height, src, optFullscreen)
	} else if isVimeo {
		fmt.Fprintf(out, _htmlBlockVideoVimeo, width, height, src)
	} else {
		var (
			optControls = ` controls`

			optAutoplay string
			optLoop     string
			optPoster   string
			withPoster  bool
		)

		optPoster, withPoster = el.Attrs[attrNamePoster]
		if withPoster {
			optPoster = fmt.Sprintf(` poster="%s"`, optPoster)
		}

		if libstrings.IsContain(el.options, optNameNocontrols) {
			optControls = ``
		}
		if libstrings.IsContain(el.options, optNameAutoplay) {
			optAutoplay = ` autoplay`
		}
		if libstrings.IsContain(el.options, optNameLoop) {
			optLoop = ` loop`
		}

		fmt.Fprintf(out, _htmlBlockVideo, src, width,
			height, optPoster, optControls, optAutoplay, optLoop)
	}

	fmt.Fprint(out, "\n</div>\n</div>")
}

func htmlWriteBody(doc *Document, out *bytes.Buffer) {
	if !doc.isEmbedded {
		fmt.Fprint(out, "\n<div id=\"content\">")
	}

	if doc.content.child != nil {
		doc.content.child.toHTML(doc, out)
	}
	if doc.content.next != nil {
		doc.content.next.toHTML(doc, out)
	}

	if !doc.isEmbedded {
		fmt.Fprint(out, "\n</div>")
	}
}

func htmlWriteFooter(doc *Document, out io.Writer) {
	fmt.Fprint(out, `
<div id="footer">
<div id="footer-text">`)

	if len(doc.Revision.Number) > 0 {
		var (
			prefix string
			ok     bool
		)

		prefix, ok = doc.Attributes[metaNameVersionLabel]
		if ok && len(prefix) == 0 {
			prefix = `Version `
		} else {
			prefix = ` `
		}

		fmt.Fprintf(out, "\n%s%s<br>", prefix, doc.Revision.Number)
	}

	if len(doc.LastUpdated) > 0 {
		fmt.Fprintf(out, "\nLast updated %s", doc.LastUpdated)
	}

	fmt.Fprint(out, "\n</div>\n</div>")
}

// htmlWriteFootnote generate HTML content for footnote.
// Each unique footnote will have its id, so it can be referenced at footer.
func htmlWriteFootnote(el *element, out io.Writer) {
	if len(el.ID) != 0 {
		// The first footnote with explicit ID.
		fmt.Fprintf(out, `<sup class="footnote" id="_footnote_%s">[<a id="_footnoteref_%d" class="footnote" href="#_footnotedef_%d" title="View footnote.">%d</a>]</sup>`,
			el.ID, el.level, el.level, el.level)

	} else if len(el.key) != 0 {
		// The first footnote without ID.
		fmt.Fprintf(out, `<sup class="footnote">[<a id="_footnoteref_%d" class="footnote" href="#_footnotedef_%d" title="View footnote.">%d</a>]</sup>`,
			el.level, el.level, el.level)
	} else {
		// The next footnote with same ID.
		fmt.Fprintf(out, `<sup class="footnoteref">[<a class="footnote" href="#_footnotedef_%d" title="View footnote.">%d</a>]</sup>`,
			el.level, el.level)
	}
}

func htmlWriteFootnoteDefs(doc *Document, out io.Writer) {
	if len(doc.footnotes) == 0 {
		return
	}

	fmt.Fprint(out, "\n")
	fmt.Fprint(out, `<div id="footnotes">`)
	fmt.Fprint(out, "\n")
	fmt.Fprint(out, `<hr>`)
	fmt.Fprint(out, "\n")

	var (
		mcr *macro
	)
	for _, mcr = range doc.footnotes {
		fmt.Fprintf(out, `<div class="footnote" id="_footnotedef_%d">`, mcr.level)
		fmt.Fprint(out, "\n")
		fmt.Fprintf(out, `<a href="#_footnoteref_%d">%d</a>. `, mcr.level, mcr.level)
		if mcr.content != nil {
			mcr.content.toHTML(doc, out)
		}
		fmt.Fprint(out, "\n")
		fmt.Fprint(out, `</div>`)
		fmt.Fprint(out, "\n")
	}
	fmt.Fprint(out, `</div>`)
	fmt.Fprint(out, "\n")
}

func htmlWriteHeader(doc *Document, out io.Writer) {
	fmt.Fprint(out, "\n<div id=\"header\">")

	var (
		author *Author
		prefix string
		sep    string
		x      int
		ok     bool
	)

	_, ok = doc.Attributes[metaNameShowTitle]
	if ok {
		_, ok = doc.Attributes[metaNameNoTitle]
		if !ok && doc.Title.el != nil {
			fmt.Fprint(out, "\n<h1>")
			doc.Title.el.toHTML(doc, out)
			fmt.Fprint(out, "</h1>")
		}
	}

	fmt.Fprint(out, "\n<div class=\"details\">")

	var authorID, emailID string
	for x, author = range doc.Authors {
		if x == 0 {
			authorID = attrValueAuthor
			emailID = attrValueEmail
		} else {
			authorID = fmt.Sprintf(`%s%d`, attrValueAuthor, x+1)
			emailID = fmt.Sprintf(`%s%d`, attrValueEmail, x+1)
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
		prefix, ok = doc.Attributes[metaNameVersionLabel]
		if ok && len(prefix) == 0 {
			prefix = defVersionPrefix
		} else {
			prefix = ` `
		}

		sep = ``
		if len(doc.Revision.Date) > 0 {
			sep = `,`
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

	if doc.tocIsEnabled && (doc.tocPosition == `` ||
		doc.tocPosition == metaValueAuto ||
		doc.tocPosition == metaValueLeft ||
		doc.tocPosition == metaValueRight) {
		doc.tocHTML(out)
	}
	fmt.Fprint(out, "\n</div>")
}

func htmlWriteInlineImage(el *element, out io.Writer) {
	var (
		classes = strings.TrimSpace(`image ` + el.htmlClasses())

		link     string
		withLink bool
	)

	fmt.Fprintf(out, `<span class=%q>`, classes)
	link, withLink = el.Attrs[attrNameLink]
	if withLink {
		fmt.Fprintf(out, `<a class=%q href=%q>`, attrValueImage, link)
	}

	var (
		src string = el.Attrs[attrNameSrc]
		alt string = el.Attrs[attrNameAlt]

		width  string
		height string
		ok     bool
	)

	width, ok = el.Attrs[attrNameWidth]
	if ok {
		width = fmt.Sprintf(` width="%s"`, width)
	}
	height, ok = el.Attrs[attrNameHeight]
	if ok {
		height = fmt.Sprintf(` height="%s"`, height)
	}

	fmt.Fprintf(out, `<img src=%q alt=%q%s%s>`, src, alt, width, height)

	if withLink {
		fmt.Fprint(out, `</a>`)
	}

	fmt.Fprint(out, `</span>`)
}

func htmlWriteListDescription(el *element, out io.Writer) {
	var openTag string
	if el.isStyleQandA() {
		htmlWriteBlockBegin(el, out, `qlist qanda`)
		openTag = "\n<ol>"
	} else if el.isStyleHorizontal() {
		htmlWriteBlockBegin(el, out, `hdlist`)
		openTag = "\n<table>"
	} else {
		htmlWriteBlockBegin(el, out, `dlist`)
		openTag = "\n<dl>"
	}

	fmt.Fprint(out, openTag)
}

func htmlWriteListDescriptionEnd(el *element, out io.Writer) {
	if el.isStyleQandA() {
		fmt.Fprintf(out, "\n</ol>\n</div>")
	} else if el.isStyleHorizontal() {
		fmt.Fprintf(out, "\n</table>\n</div>")
	} else {
		fmt.Fprintf(out, "\n</dl>\n</div>")
	}
}

func htmlWriteListOrdered(el *element, out io.Writer) {
	var (
		class string = el.getListOrderedClass()
		tipe  string = el.getListOrderedType()
	)

	if len(tipe) > 0 {
		tipe = ` type="` + tipe + `"`
	}

	htmlWriteBlockBegin(el, out, "olist "+class)

	fmt.Fprintf(out, "\n<ol class=\"%s\"%s>", class, tipe)
}

func htmlWriteListOrderedEnd(out io.Writer) {
	fmt.Fprint(out, "\n</ol>\n</div>")
}

func htmlWriteListUnordered(el *element, out io.Writer) {
	var classes string
	if len(el.rawStyle) != 0 {
		classes = fmt.Sprintf(" class=%q", el.rawStyle)
	}
	htmlWriteBlockBegin(el, out, "")
	fmt.Fprintf(out, "\n<ul%s>", classes)
}

func htmlWriteListUnorderedEnd(out io.Writer) {
	fmt.Fprint(out, "\n</ul>\n</div>")
}

func htmlWriteParagraphBegin(el *element, out io.Writer) {
	htmlWriteBlockBegin(el, out, "paragraph")
	fmt.Fprint(out, "\n<p>")
}

func htmlWriteSection(doc *Document, el *element, out io.Writer) {
	var (
		class string
		tag   string
	)

	switch el.kind {
	case elKindSectionL1:
		class = "sect1"
		tag = "h2"
	case elKindSectionL2:
		class = "sect2"
		tag = "h3"
	case elKindSectionL3:
		class = "sect3"
		tag = "h4"
	case elKindSectionL4:
		class = "sect4"
		tag = "h5"
	case elKindSectionL5:
		class = "sect5"
		tag = "h6"
	}

	fmt.Fprintf(out, "\n<div class=%q>\n<%s id=%q>", class, tag, el.ID)

	var (
		withSectAnchors bool
		withSectlinks   bool
	)

	_, withSectAnchors = doc.Attributes[metaNameSectAnchors]
	if withSectAnchors {
		fmt.Fprintf(out, `<a class="anchor" href="#%s"></a>`, el.ID)
	}
	_, withSectlinks = doc.Attributes[metaNameSectLinks]
	if withSectlinks {
		fmt.Fprintf(out, `<a class="link" href="#%s">`, el.ID)
	}

	if el.sectnums != nil && el.level <= doc.sectLevel {
		fmt.Fprint(out, el.sectnums.String())
	}

	el.title.toHTML(doc, out)

	if withSectlinks {
		fmt.Fprint(out, "</a>")
	}
	fmt.Fprintf(out, "</%s>", tag)

	if el.kind == elKindSectionL1 {
		fmt.Fprint(out, "\n<div class=\"sectionbody\">")
	}
}

func hmltWriteSectionDiscrete(doc *Document, el *element, out io.Writer) {
	var (
		tag string
	)
	switch el.level {
	case elKindSectionL1:
		tag = "h2"
	case elKindSectionL2:
		tag = "h3"
	case elKindSectionL3:
		tag = "h4"
	case elKindSectionL4:
		tag = "h5"
	case elKindSectionL5:
		tag = "h6"
	}

	fmt.Fprintf(out, "\n<%s id=%q class=%q>", tag, el.ID, attrNameDiscrete)
	el.title.toHTML(doc, out)
	fmt.Fprintf(out, "</%s>", tag)
}

func htmlWriteTable(doc *Document, el *element, out io.Writer) {
	var (
		table *elementTable = el.table

		footer *tableRow
		format *columnFormat
		style  string

		withTableCaption bool
	)

	if table == nil {
		return
	}

	fmt.Fprintf(out, "\n<table class=%q", table.classes.String())

	style = table.htmlStyle()
	if len(style) > 0 {
		fmt.Fprintf(out, " style=%q", style)
	}
	fmt.Fprint(out, ">")

	if len(el.rawTitle) > 0 {
		var (
			caption string
			ok      bool
		)

		doc.counterTable++
		_, withTableCaption = doc.Attributes[metaNameTableCaption]

		if withTableCaption {
			caption, ok = el.Attrs[attrNameCaption]
			if !ok {
				caption = fmt.Sprintf("Table %d.", doc.counterTable)
			}
		}
		fmt.Fprintf(out, "\n<caption class=%q>%s %s</caption>",
			attrValueTitle, caption, el.rawTitle)
	}

	fmt.Fprint(out, "\n<colgroup>")
	for _, format = range table.formats {
		if format.width != nil {
			fmt.Fprintf(out, "\n<col style=\"width: %s%%;\">", format.width)
		} else {
			fmt.Fprint(out, "\n<col>")
		}
	}
	fmt.Fprint(out, "\n</colgroup>")

	var (
		rows []*tableRow = table.rows
		row  *tableRow
	)

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
		for _, row = range rows {
			htmlWriteTableRow(doc, table, row, out)
		}
		fmt.Fprint(out, "\n</tbody>")
	}

	if table.hasFooter && footer != nil {
		htmlWriteTableFooter(doc, table, footer, out)
	}

	fmt.Fprint(out, "\n</table>")
}

func htmlWriteTableFooter(doc *Document, table *elementTable, footer *tableRow, out io.Writer) {
	fmt.Fprint(out, "\n<tfoot>")
	htmlWriteTableRow(doc, table, footer, out)
	fmt.Fprint(out, "\n</tfoot>")

}

func htmlWriteTableHeader(doc *Document, header *tableRow, out io.Writer) {
	var (
		classRow = "tableblock halign-left valign-top"

		cell *tableCell
		cont *element
	)

	fmt.Fprint(out, "\n<thead>\n<tr>")
	for _, cell = range header.cells {
		fmt.Fprintf(out, "\n<th class=%q>", classRow)
		cont = parseInlineMarkup(doc, bytes.TrimSpace(cell.content))
		cont.toHTML(doc, out)
		fmt.Fprint(out, "</th>")
	}
	fmt.Fprint(out, "\n</tr>\n</thead>")
}

func htmlWriteTableRow(doc *Document, table *elementTable, row *tableRow, out io.Writer) {
	var (
		cell      *tableCell
		format    *columnFormat
		subdoc    *Document
		container *element

		tag     string
		colspan string

		contentTrimmed []byte
		rawParagraphs  [][]byte
		p              []byte
		x              int
	)

	fmt.Fprint(out, "\n<tr>")
	for x, cell = range row.cells {
		format = table.formats[x]
		tag = "td"
		colspan = ""

		if format.style == colStyleHeader {
			tag = "th"
		}
		if cell.format.nspanCol > 0 {
			colspan = fmt.Sprintf(` colspan="%d"`, cell.format.nspanCol)
		}

		fmt.Fprintf(out, "\n<%s class=%q%s>", tag,
			format.htmlClasses(), colspan)

		contentTrimmed = bytes.TrimSpace(cell.content)

		switch format.style {
		case colStyleAsciidoc:
			subdoc = parseSub(doc, contentTrimmed)
			fmt.Fprint(out, "\n<div id=\"content\">")
			_ = subdoc.ToHTMLEmbedded(out)
			fmt.Fprint(out, "\n</div>")

		case colStyleDefault:
			rawParagraphs = bytes.Split(contentTrimmed, []byte("\n\n"))
			for x, p = range rawParagraphs {
				if x > 0 {
					fmt.Fprint(out, "\n")
				}
				fmt.Fprintf(out, "<p class=%q>", classNameTableBlock)
				container = parseInlineMarkup(doc, p)
				container.toHTML(doc, out)
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

func htmlWriteToC(doc *Document, el *element, out io.Writer, level int) {
	var (
		isDiscrete = el.style&styleSectionDiscrete > 0

		sectClass string
		n         int
	)

	switch el.kind {
	case elKindSectionL1:
		sectClass = "sectlevel1"
	case elKindSectionL2:
		sectClass = "sectlevel2"
	case elKindSectionL3:
		sectClass = "sectlevel3"
	case elKindSectionL4:
		sectClass = "sectlevel4"
	case elKindSectionL5:
		sectClass = "sectlevel5"
	}
	if el.level > doc.TOCLevel {
		sectClass = ""
	}

	if len(sectClass) > 0 && !isDiscrete {
		if level < el.level {
			fmt.Fprintf(out, "\n<ul class=\"%s\">", sectClass)
		} else if level > el.level {
			n = level
			for n > el.level {
				fmt.Fprint(out, "\n</ul>")
				n--
			}
		}

		fmt.Fprintf(out, "\n<li><a href=\"#%s\">", el.ID)

		if el.sectnums != nil {
			fmt.Fprint(out, el.sectnums.String())
		}

		doc.isForToC = true
		el.title.toHTML(doc, out)
		doc.isForToC = false

		fmt.Fprint(out, "</a>")
	}

	if el.child != nil {
		htmlWriteToC(doc, el.child, out, el.level)
	}
	if len(sectClass) > 0 && !isDiscrete {
		fmt.Fprint(out, "</li>")
	}
	if el.next != nil {
		htmlWriteToC(doc, el.next, out, el.level)
	}

	if len(sectClass) > 0 && level < el.level {
		fmt.Fprint(out, "\n</ul>\n")
	}
}

func htmlWriteURLBegin(el *element, out io.Writer) {
	fmt.Fprintf(out, "<a href=\"%s\"", el.Attrs[attrNameHref])

	var (
		classes string = el.htmlClasses()
		target  string = el.Attrs[attrNameTarget]
		rel     string = el.Attrs[attrNameRel]
	)

	if len(classes) > 0 {
		fmt.Fprintf(out, ` class="%s"`, classes)
	}
	if len(target) > 0 {
		fmt.Fprintf(out, ` target="%s"`, target)
	}
	if len(rel) > 0 {
		fmt.Fprintf(out, ` rel="%s"`, rel)
	}
	fmt.Fprintf(out, `>%s`, el.raw)
}

func htmlWriteURLEnd(out io.Writer) {
	fmt.Fprint(out, "</a>")
}
