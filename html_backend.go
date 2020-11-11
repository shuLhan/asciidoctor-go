// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
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

func htmlWriteBlockBegin(node *adocNode, out io.Writer, addClass string) (err error) {
	_, err = fmt.Fprint(out, "\n<div")
	if err != nil {
		return err
	}
	if len(node.ID) > 0 {
		_, err = fmt.Fprintf(out, ` id="%s"`, node.ID)
		if err != nil {
			return err
		}
	}

	classes := node.Classes()
	c := strings.TrimSpace(addClass + " " + classes)
	if len(c) > 0 {
		_, err = fmt.Fprintf(out, ` class="%s">`, c)
	} else {
		_, err = fmt.Fprint(out, ">")
	}

	if !(node.IsStyleAdmonition() ||
		node.kind == nodeKindBlockImage ||
		node.kind == nodeKindBlockExample ||
		node.kind == nodeKindBlockSidebar) &&
		len(node.rawTitle) > 0 {

		_, err = fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
		if err != nil {
			return err
		}
	}

	return err
}

func htmlWriteBlockAdmonition(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "admonitionblock")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, "\n<table>\n<tr>\n<td class=\"icon\">")
	if err != nil {
		return err
	}

	iconsFont := node.Attrs[attrNameIcons]
	if iconsFont == attrValueFont {
		_, err = fmt.Fprintf(out, _htmlAdmonitionIconsFont,
			strings.ToLower(node.Classes()), node.rawLabel.String())
	} else {
		_, err = fmt.Fprintf(out, _htmlBlockTitle,
			node.rawLabel.String())
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, _htmlAdmonitionContent, node.raw)
	if err != nil {
		return err
	}

	if len(node.rawTitle) > 0 {
		_, err = fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
	}

	return err
}

func htmlWriteBlockAudio(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "audioblock")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, _htmlBlockContent)
	if err != nil {
		return err
	}

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

	_, err = fmt.Fprintf(out, _htmlBlockAudio, src, optAutoplay,
		optControls, optLoop)

	return err
}

func htmlWriteBlockExample(doc *Document, node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "exampleblock")
	if err != nil {
		return err
	}
	if len(node.rawTitle) > 0 {
		doc.counterExample++
		_, err = fmt.Fprintf(out, _htmlBlockExampleTitle,
			doc.counterExample, node.rawTitle)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(out, _htmlBlockContent)
	if err != nil {
		return err
	}
	return nil
}

func htmlWriteBlockImage(doc *Document, node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "imageblock")
	if err != nil {
		return err
	}

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

	_, err = fmt.Fprintf(out, _htmlBlockImage, src, alt, width, height)
	if err != nil {
		return err
	}

	if len(node.rawTitle) > 0 {
		doc.counterImage++
		_, err = fmt.Fprintf(out, _htmlBlockImageTitle,
			doc.counterImage, node.rawTitle)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(out, _htmlBlockImageEnd)

	return err
}

func htmlWriteBlockLiteral(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, _htmlBlockLiteralContent, node.raw)
	if err != nil {
		return err
	}

	return err
}

func htmlWriteBlockOpenBegin(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "openblock")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, _htmlBlockContent)

	return err
}

func htmlWriteBlockQuote(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "quoteblock")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, _htmlBlockQuoteBegin, node.raw)

	return err
}

func htmlWriteBlockQuoteEnd(node *adocNode, out io.Writer) (err error) {
	_, err = fmt.Fprint(out, _htmlBlockQuoteEnd)
	if err != nil {
		return err
	}
	if len(node.key) > 0 {
		_, err = fmt.Fprintf(out, _htmlQuoteAuthor, node.key)
		if err != nil {
			return err
		}
	}
	if len(node.value) > 0 {
		_, err = fmt.Fprintf(out, _htmlQuoteCitation, node.value)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(out, _htmlBlockEnd)

	return err
}

func htmlWriteBlockSidebar(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "sidebarblock")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, _htmlBlockContent)
	if err != nil {
		return err
	}
	if len(node.rawTitle) > 0 {
		_, err = fmt.Fprintf(out, _htmlBlockTitle, node.rawTitle)
	}
	return err
}

func htmlWriteBlockVerse(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "verseblock")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, _htmlBlockVerse, node.raw)
	if err != nil {
		return err
	}
	return nil
}

func htmlWriteBlockVerseEnd(node *adocNode, out io.Writer) (err error) {
	_, err = fmt.Fprint(out, _htmlBlockVerseEnd)
	if err != nil {
		return err
	}
	if len(node.key) > 0 {
		_, err = fmt.Fprintf(out, _htmlQuoteAuthor, node.key)
		if err != nil {
			return err
		}
	}
	if len(node.value) > 0 {
		_, err = fmt.Fprintf(out, _htmlQuoteCitation, node.value)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(out, _htmlBlockEnd)

	return err
}

func htmlWriteBlockVideo(node *adocNode, out io.Writer) (err error) {
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

	err = htmlWriteBlockBegin(node, out, "videoblock")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, _htmlBlockContent)
	if err != nil {
		return err
	}

	if isYoutube {
		optFullscreen, noFullscreen := node.Attrs[optVideoNofullscreen]
		if !noFullscreen {
			optFullscreen = " allowfullscreen"
		}
		_, err = fmt.Fprintf(out, _htmlBlockVideoYoutube, width,
			height, src, optFullscreen)
	} else if isVimeo {
		_, err = fmt.Fprintf(out, _htmlBlockVideoVimeo, width,
			height, src)
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

		_, err = fmt.Fprintf(out, _htmlBlockVideo, src, width,
			height, optPoster, optControls, optAutoplay, optLoop)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, _htmlBlockEnd)

	return err
}

func htmlWriteBody(doc *Document, out io.Writer) (err error) {
	err = htmlWriteHeader(doc, out)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, _htmlContentBegin)
	if err != nil {
		return err
	}

	if doc.content.child != nil {
		err = doc.content.child.toHTML(doc, out, false)
		if err != nil {
			return err
		}
	}
	if doc.content.next != nil {
		err = doc.content.next.toHTML(doc, out, false)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(out, _htmlContentEnd)
	if err != nil {
		return err
	}

	return nil
}

func htmlWriteHeader(doc *Document, out io.Writer) (
	err error,
) {
	_, err = fmt.Fprint(out, _htmlHeaderBegin)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, _htmlHeaderTitleBegin)
	if err != nil {
		return err
	}
	err = doc.title.toHTML(doc, out, false)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, _htmlHeaderTitleEnd)
	if err != nil {
		return err
	}

	if _, err = fmt.Fprint(out, _htmlHeaderDetail); err != nil {
		return err
	}
	if len(doc.Author) > 0 {
		_, err = fmt.Fprintf(out, _htmlHeaderDetailAuthor, doc.Author)
		if err != nil {
			return err
		}
	}
	if len(doc.RevNumber) > 0 {
		_, err = fmt.Fprintf(out, _htmlHeaderDetailRevNumber,
			doc.RevNumber, doc.RevSeparator)
		if err != nil {
			return err
		}
	}
	if len(doc.RevDate) > 0 {
		_, err = fmt.Fprintf(out, _htmlHeaderDetailRevDate, doc.RevDate)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(out, _htmlHeaderDetailEnd)
	if err != nil {
		return err
	}

	if doc.tocIsEnabled && (doc.tocPosition == "" ||
		doc.tocPosition == metaValueAuto ||
		doc.tocPosition == metaValueLeft ||
		doc.tocPosition == metaValueRight) {
		err = doc.tocHTML(out)
		if err != nil {
			return fmt.Errorf("ToHTML: %w", err)
		}
	}

	_, err = fmt.Fprint(out, _htmlHeaderEnd)
	if err != nil {
		return err
	}

	return nil
}

func htmlWriteInlineImage(node *adocNode, out io.Writer) (err error) {
	classes := strings.TrimSpace("image " + node.Classes())
	_, err = fmt.Fprintf(out, _htmlInlineImage, classes)
	if err != nil {
		return fmt.Errorf("htmlWriteInlineImage: %w", err)
	}
	link, withLink := node.Attrs[attrNameLink]
	if withLink {
		_, err = fmt.Fprintf(out, _htmlInlineImageLink, link)
		if err != nil {
			return err
		}
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

	_, err = fmt.Fprintf(out, _htmlInlineImageImage, src, alt, width, height)
	if err != nil {
		return err
	}

	if withLink {
		_, err = fmt.Fprint(out, `</a>`)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(out, `</span>`)

	return err
}

func htmlWriteListDescription(node *adocNode, out io.Writer) (err error) {
	var openTag string
	if node.IsStyleQandA() {
		err = htmlWriteBlockBegin(node, out, "qlist qanda")
		openTag = "\n<ol>"
	} else if node.IsStyleHorizontal() {
		err = htmlWriteBlockBegin(node, out, "hdlist")
		openTag = "\n<table>"
	} else {
		err = htmlWriteBlockBegin(node, out, "dlist")
		openTag = "\n<dl>"
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, openTag)

	return err
}

func htmlWriteListDescriptionEnd(node *adocNode, out io.Writer) (err error) {
	if node.IsStyleQandA() {
		_, err = fmt.Fprintf(out, "\n</ol>\n</div>")
	} else if node.IsStyleHorizontal() {
		_, err = fmt.Fprintf(out, "\n</table>\n</div>")
	} else {
		_, err = fmt.Fprintf(out, "\n</dl>\n</div>")
	}
	return err
}

func htmlWriteListOrdered(node *adocNode, out io.Writer) (err error) {
	class := node.getListOrderedClass()
	tipe := node.getListOrderedType()
	if len(tipe) > 0 {
		tipe = ` type="` + tipe + `"`
	}

	err = htmlWriteBlockBegin(node, out, "olist "+class)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "\n<ol class=\"%s\"%s>", class, tipe)

	return err
}

func htmlWriteListOrderedEnd(out io.Writer) (err error) {
	_, err = fmt.Fprint(out, "\n</ol>\n</div>")
	return err
}

func htmlWriteListUnordered(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "\n<ul>")

	return err
}

func htmlWriteListUnorderedEnd(out io.Writer) (err error) {
	_, err = fmt.Fprint(out, "\n</ul>\n</div>")
	return err
}

func htmlWriteParagraphBegin(node *adocNode, out io.Writer) (err error) {
	err = htmlWriteBlockBegin(node, out, "paragraph")
	if err != nil {
		return err
	}

	_, err = out.Write([]byte("\n<p>"))

	return err
}

func htmlWriteSection(doc *Document, node *adocNode, out io.Writer, isForToC bool) (
	err error,
) {
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

	_, err = fmt.Fprintf(out, _htmlSection, class, tag, node.ID)
	if err != nil {
		return err
	}

	if node.sectnums != nil && node.level <= doc.sectLevel {
		_, err = out.Write([]byte(node.sectnums.String()))
		if err != nil {
			return err
		}
	}

	err = node.title.toHTML(doc, out, isForToC)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "</%s>", tag)
	if err != nil {
		return err
	}

	if node.kind == nodeKindSectionL1 {
		_, err = out.Write([]byte("\n<div class=\"sectionbody\">"))
		if err != nil {
			return err
		}
	}

	return nil
}

func htmlWriteToC(
	doc *Document, node *adocNode, out io.Writer, level int,
) (err error) {
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
			_, err = fmt.Fprintf(out, "\n<ul class=\"%s\">", sectClass)
		} else if level > node.level {
			n := level
			for n > node.level {
				_, err = out.Write([]byte("\n</ul>"))
				n--
			}
		}

		_, err = fmt.Fprintf(out, "\n<li><a href=\"#%s\">", node.ID)
		if err != nil {
			return fmt.Errorf("htmlWriteToC: %w", err)
		}

		if node.sectnums != nil {
			_, err = out.Write([]byte(node.sectnums.String()))
			if err != nil {
				return fmt.Errorf("htmlWriteToC: %w", err)
			}
		}

		err = node.title.toHTML(doc, out, true)
		if err != nil {
			return fmt.Errorf("htmlWriteToC: %w", err)
		}

		_, err = out.Write([]byte("</a>"))
		if err != nil {
			return fmt.Errorf("htmlWriteToC: %w", err)
		}
	}

	if node.child != nil {
		err = htmlWriteToC(doc, node.child, out, node.level)
		if err != nil {
			return err
		}
	}
	if len(sectClass) > 0 {
		_, err = out.Write([]byte("</li>"))
		if err != nil {
			return fmt.Errorf("htmlWriteToC: %w", err)
		}
	}
	if node.next != nil {
		err = htmlWriteToC(doc, node.next, out, node.level)
		if err != nil {
			return err
		}
	}

	if len(sectClass) > 0 && level < node.level {
		_, err = out.Write([]byte("\n</ul>\n"))
		if err != nil {
			return fmt.Errorf("htmlWriteToC: %w", err)
		}
	}

	return nil
}

func htmlWriteURLBegin(node *adocNode, out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "<a href=\"%s\"", node.Attrs[attrNameHref])
	if err != nil {
		return err
	}
	classes := node.Classes()
	if len(classes) > 0 {
		_, err = fmt.Fprintf(out, ` class="%s"`, classes)
		if err != nil {
			return err
		}
	}
	target := node.Attrs[attrNameTarget]
	if len(target) > 0 {
		_, err = fmt.Fprintf(out, ` target="%s"`, target)
		if err != nil {
			return err
		}
	}
	rel := node.Attrs[attrNameRel]
	if len(rel) > 0 {
		_, err = fmt.Fprintf(out, ` rel="%s"`, rel)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, `>%s`, node.raw)

	return err
}

func htmlWriteURLEnd(out io.Writer) (err error) {
	_, err = fmt.Fprint(out, "</a>")
	return err
}
