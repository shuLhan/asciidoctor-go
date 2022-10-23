// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

// toRoff convert the element content to ROFF format.
func (el *element) toRoff(out io.Writer, doc *Document) {
	el.toRoffBegin(out, doc)

	if el.child != nil {
		el.child.toRoff(out, doc)
	}

	el.toRoffEnd(out, doc)

	if el.next != nil {
		el.next.toRoff(out, doc)
	}
}

func (el *element) toRoffBegin(out io.Writer, doc *Document) {
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
		// fmt.Fprintf(out, `<a href="#%s">%s</a>`, href, label)

	case elKindFootnote:
		// htmlWriteFootnote(el, w)

	case elKindMacroTOC:
		// N/A.

	case elKindPreamble:
		// ???

	case elKindSectionDiscrete:
		// htmlWriteSectionDiscrete(doc, el, w)

	case elKindSectionL1, elKindSectionL2, elKindSectionL3, elKindSectionL4, elKindSectionL5:
		roffWriteSection(out, doc, el)

	case elKindParagraph:
		if el.isStyleAdmonition() {
			//htmlWriteBlockAdmonition(el, w)
		} else if el.isStyleQuote() {
			//htmlWriteBlockQuote(el, w)
		} else if el.isStyleVerse() {
			//htmlWriteBlockVerse(el, w)
		} else {
			roffWriteParagraphBegin(out, el)
		}

	case elKindLiteralParagraph, elKindBlockLiteral, elKindBlockLiteralNamed, elKindBlockListing, elKindBlockListingNamed:
		roffWriteBlockLiteral(out, el)

	case elKindInlineImage:
		//htmlWriteInlineImage(el, w)

	case elKindInlinePass:
		//htmlWriteInlinePass(doc, el, w)

	case elKindListDescription:
		//htmlWriteListDescription(el, w)
	case elKindListOrdered:
		//htmlWriteListOrdered(el, w)
	case elKindListUnordered:
		//htmlWriteListUnordered(el, w)

	case elKindListOrderedItem:
		roffWriteListOrderedItem(out, el)

	case elKindListUnorderedItem:
		roffWriteListUnorderedItem(out, el)

	case elKindListDescriptionItem:
		roffWriteListDescriptionItem(out, doc, el)

	case lineKindHorizontalRule:
		//fmt.Fprint(w, "\n<hr>")

	case lineKindPageBreak:
		//fmt.Fprint(w, "\n<div style=\"page-break-after: always;\"></div>")

	case elKindBlockExample:
		if el.isStyleAdmonition() {
			//htmlWriteBlockAdmonition(el, w)
		} else {
			//htmlWriteBlockExample(doc, el, w)
		}

	case elKindBlockImage:
		//htmlWriteBlockImage(doc, el, w)

	case elKindBlockOpen:
		if el.isStyleAdmonition() {
			//htmlWriteBlockAdmonition(el, w)
		} else if el.isStyleQuote() {
			//htmlWriteBlockQuote(el, w)
		} else if el.isStyleVerse() {
			//htmlWriteBlockVerse(el, w)
		} else {
			//fmt.Fprint(out, _roffTmplBlockOpenBegin)
		}

	case elKindBlockPassthrough:
		fmt.Fprintf(out, _lf+`%s`, el.raw)

	case elKindBlockExcerpts:
		if el.isStyleVerse() {
			//htmlWriteBlockVerse(el, w)
		} else {
			//htmlWriteBlockQuote(el, w)
		}

	case elKindBlockSidebar:
		//htmlWriteBlockSidebar(el, w)

	case elKindBlockVideo:
		//htmlWriteBlockVideo(el, w)

	case elKindBlockAudio:
		//htmlWriteBlockAudio(el, w)

	case elKindInlineID:
		if !doc.isForToC {
			//fmt.Fprintf(w, "<a id=%q></a>", el.ID)
		}

	case elKindInlineIDShort:
		if !doc.isForToC {
			//fmt.Fprintf(w, "<span id=%q>%s", el.ID, el.raw)
		}

	case elKindInlineParagraph:
		roffWriteText(out, el.raw)

	case elKindPassthrough:
		fmt.Fprint(out, string(el.raw))
	case elKindPassthroughDouble:
		fmt.Fprint(out, string(el.raw))
	case elKindPassthroughTriple:
		fmt.Fprint(out, string(el.raw))

	case elKindSymbolQuoteDoubleBegin:
		//fmt.Fprint(out, symbolQuoteDoubleBegin, string(el.raw))
	case elKindSymbolQuoteDoubleEnd:
		//fmt.Fprint(out, symbolQuoteDoubleEnd, string(el.raw))

	case elKindSymbolQuoteSingleBegin:
		//fmt.Fprint(w, symbolQuoteSingleBegin, string(el.raw))
	case elKindSymbolQuoteSingleEnd:
		//fmt.Fprint(w, symbolQuoteSingleEnd, string(el.raw))

	case elKindText:
		roffWriteText(out, el.raw)

	case elKindTextBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(out, `\fB`)
		} else if len(el.raw) > 0 {
			fmt.Fprint(out, `*`)
		}
		roffWriteText(out, el.raw)

	case elKindUnconstrainedBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(out, `\fB`)
		} else if len(el.raw) > 0 {
			fmt.Fprint(out, `**`)
		}
		roffWriteText(out, el.raw)

	case elKindTextItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(out, `\fI`)
		} else if len(el.raw) > 0 {
			fmt.Fprint(out, `_`)
		}
		roffWriteText(out, el.raw)

	case elKindUnconstrainedItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(out, `\fI`)
		} else if len(el.raw) > 0 {
			fmt.Fprint(out, `__`)
		}
		roffWriteText(out, el.raw)

	case elKindTextMono:
		if el.hasStyle(styleTextMono) {
			//fmt.Fprint(w, "<code>")
		} else if len(el.raw) > 0 {
			//fmt.Fprint(w, "`")
		}
		fmt.Fprint(out, string(el.raw))

	case elKindUnconstrainedMono:
		if el.hasStyle(styleTextMono) {
			//fmt.Fprint(w, "<code>")
		} else if len(el.raw) > 0 {
			//fmt.Fprint(w, "``")
		}
		fmt.Fprint(out, string(el.raw))

	case elKindURL:
		roffWriteUrl(out, doc, el)

	case elKindTextSubscript:
		//fmt.Fprintf(w, "<sub>%s</sub>", el.raw)
	case elKindTextSuperscript:
		//fmt.Fprintf(w, "<sup>%s</sup>", el.raw)

	case elKindTable:
		//htmlWriteTable(doc, el, w)
	}
}

func (el *element) toRoffEnd(out io.Writer, doc *Document) {
	switch el.kind {
	case elKindPreamble:
		if doc.tocIsEnabled && doc.tocPosition == metaValuePreamble {
			//doc.tocHTML(w)
		}

	case elKindParagraph:
		if el.isStyleAdmonition() {
			//fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if el.isStyleQuote() {
			//htmlWriteBlockQuoteEnd(el, w)
		} else if el.isStyleVerse() {
			//htmlWriteBlockVerseEnd(el, w)
		} else {
			fmt.Fprint(out, _lf)
		}

	case elKindListOrderedItem:
		fmt.Fprint(out, _roffTmplListOrderedItemEnd)

	case elKindListUnorderedItem:
		fmt.Fprint(out, _roffTmplListUnorderedItemEnd)

	case elKindListDescriptionItem:
		fmt.Fprint(out, _roffTmplListDescriptionEnd)

	case elKindListDescription:
		//htmlWriteListDescriptionEnd(el, w)
	case elKindListOrdered:
		//htmlWriteListOrderedEnd(w)
	case elKindListUnordered:
		//htmlWriteListUnorderedEnd(w)

	case elKindBlockExample:
		if el.isStyleAdmonition() {
			//fmt.Fprint(w, _htmlAdmonitionEnd)
		} else {
			//fmt.Fprint(w, "\n</div>\n</div>")
		}

	case elKindBlockOpen:
		if el.isStyleAdmonition() {
			//fmt.Fprint(w, _htmlAdmonitionEnd)
		} else if el.isStyleQuote() {
			//htmlWriteBlockQuoteEnd(el, w)
		} else if el.isStyleVerse() {
			//htmlWriteBlockVerseEnd(el, w)
		} else {
			//fmt.Fprint(out, _roffTmplBlockOpenEnd)
		}
	case elKindBlockExcerpts:
		if el.isStyleVerse() {
			//htmlWriteBlockVerseEnd(el, w)
		} else {
			//htmlWriteBlockQuoteEnd(el, w)
		}

	case elKindBlockSidebar:
		//fmt.Fprint(w, "\n</div>\n</div>")

	case elKindInlineIDShort:
		if !doc.isForToC {
			//fmt.Fprint(w, "</span>")
		}

	case elKindInlineParagraph:
		fmt.Fprint(out, _lf)

	case elKindTextBold, elKindUnconstrainedBold:
		if el.hasStyle(styleTextBold) {
			fmt.Fprint(out, `\fP`)
		}
	case elKindTextItalic, elKindUnconstrainedItalic:
		if el.hasStyle(styleTextItalic) {
			fmt.Fprint(out, `\fP`)
		}
	case elKindTextMono, elKindUnconstrainedMono:
		if el.hasStyle(styleTextMono) {
			//fmt.Fprint(w, "</code>")
		}
	}
}

// roffEscape escape character '-', '\' from input using backslash.
func roffEscape(input []byte) (out []byte) {
	var (
		c byte
	)

	out = make([]byte, 0, len(input))
	for _, c = range input {
		if c == '-' {
			out = append(out, '\\', c)
		} else if c == '\\' {
			out = append(out, '\\', '(', 'r', 's')
		} else {
			out = append(out, c)
		}
	}
	return out
}

func roffEscapeString(input string) (out string) {
	var (
		outb = roffEscape([]byte(input))
	)
	return string(outb)
}

func roffWriteHeader(out io.Writer, doc *Document) {
	var (
		generator    = doc.Attributes[MetaNameGenerator]
		manLinkstyle = doc.Attributes[MetanameManLinkstyle]
		manManual    = doc.Attributes[MetanameManManual]
		manSource    = doc.Attributes[MetanameManSource]
		manTitle     = doc.Attributes[MetanameManTitle]
		manVolnum    = doc.Attributes[MetanameManVolnum]
		manTitleUp   = strings.ToUpper(manTitle)
	)

	fmt.Fprintln(out, `'\" t`)
	fmt.Fprintf(out, `.\"     Title: %s`+_lf, manTitle)

	if len(doc.Authors) > 0 {
		fmt.Fprintf(out, `.\"    Author: %s`+_lf, doc.Authors[0].FullName())
	}

	fmt.Fprintf(out, `.\" Generator: %s`+_lf, generator)
	fmt.Fprintf(out, `.\"      Date: %s`+_lf, doc.Revision.Date)
	fmt.Fprintf(out, `.\"    Manual: %s`+_lf, manManual)
	fmt.Fprintf(out, `.\"    Source: %s`+_lf, manSource)

	fmt.Fprintln(out, `.\"  Language: English`)
	fmt.Fprintln(out, `.\"`)

	fmt.Fprintf(out, `.TH "%s" "%s" "%s" "%s" "%s"`+_lf,
		roffEscapeString(manTitleUp), manVolnum, doc.Revision.Date,
		roffEscapeString(manSource), roffEscapeString(manManual))

	// Set register Aq to \(aq if run using groff; else '.
	fmt.Fprintln(out, `.ie \n(.g .ds Aq \(aq`)
	fmt.Fprintln(out, `.el       .ds Aq '`)

	// Set size of a space between words to value of register .ss and a
	// space between sentence to 0.
	fmt.Fprintln(out, `.ss \n[.ss] 0`)

	// Disable hyphenation.
	fmt.Fprintln(out, `.nh`)

	// Adjust text to the left margin.
	fmt.Fprintln(out, `.ad l`)

	// Define a macro named URL with three parameters.
	// The second parameter is in italic, the first parameter wrapped
	// with "< >" and the third parameter in the last.
	fmt.Fprintln(out, `.de URL`+_lf+`\fI\\$2\fP <\\$1>\\$3`+_lf+`..`)

	// Create alias MTO for macro URL.
	fmt.Fprintln(out, `.als MTO URL`)

	// If run under groff, include www.tmac, extend the macro URL and MTO
	// with ".ad l"; and set register LINKSTYLE.
	fmt.Fprintf(out, _roffTmplMacroLink, manLinkstyle)
}

func roffWriteBody(out io.Writer, doc *Document) {
	if doc.content.child != nil {
		doc.content.child.toRoff(out, doc)
	}
	if doc.content.next != nil {
		doc.content.next.toRoff(out, doc)
	}
}

func roffWriteFooter(out io.Writer, doc *Document) {
	var (
		author *Author
		x      int
	)
	fmt.Fprint(out, _roffTmplSectionAuthor)
	for x, author = range doc.Authors {
		if x != 0 {
			fmt.Fprint(out, `.sp`+_lf)
		}
		fmt.Fprint(out, author.FullName()+_lf)
	}
}

func roffWriteBlockLiteral(out io.Writer, el *element) {
	fmt.Fprint(out, _roffTmplBlockLiteralBegin)
	roffWriteText(out, el.raw)
	fmt.Fprint(out, _lf)
	fmt.Fprint(out, _roffTmplBlockLiteralEnd)
}

func roffWriteListOrderedItem(out io.Writer, el *element) {
	if el.prev != nil && el.prev.kind == elKindListOrderedItem {
		fmt.Fprint(out, `.sp`+_lf)
	} else {
		var parent = el.parent // parent is elKindListOrdered.
		if parent != nil {
			parent = parent.parent
			if parent != nil && !parent.isSection() {
				// If the parent is section header, skip adding space.
				fmt.Fprint(out, `.sp`+_lf)
			}
		}
	}
	fmt.Fprintf(out, _roffTmplListOrderedItemBegin, el.listItemNumber, el.listItemNumber)
}

func roffWriteListUnorderedItem(out io.Writer, el *element) {
	if el.prev != nil && el.prev.kind == elKindListUnorderedItem {
		fmt.Fprint(out, `.sp`+_lf)
	} else {
		var parent = el.parent // parent is elKindListUnordered.
		if parent != nil {
			parent = parent.parent
			if parent != nil && !parent.isSection() {
				// If the parent is section header, skip adding space.
				fmt.Fprint(out, `.sp`+_lf)
			}
		}
	}
	fmt.Fprint(out, _roffTmplListUnorderedItemBegin)
}

func roffWriteListDescriptionItem(out io.Writer, doc *Document, el *element) {
	if el.prev != nil && el.prev.kind == elKindListDescriptionItem {
		fmt.Fprint(out, `.sp`+_lf)
	} else {
		var parent = el.parent // parent is elKindListDescription.
		if parent != nil {
			parent = parent.parent
			if parent != nil && !parent.isSection() {
				// If the parent is section header, skip adding space.
				fmt.Fprint(out, `.sp`+_lf)
			}
		}
	}
	var label bytes.Buffer
	if el.label != nil {
		el.label.toRoff(&label, doc)
	} else {
		label.Write(el.rawLabel.Bytes())
	}
	fmt.Fprintf(out, _roffTmplListDescriptionBegin, label.String())
}

func roffWriteParagraphBegin(out io.Writer, el *element) {
	if el.prev == nil && el.parent != nil && el.parent.isSection() {
		// Prevent double .sp if previous element is section.
		return
	}
	fmt.Fprint(out, `.sp`+_lf)
}

func roffWriteSection(out io.Writer, doc *Document, el *element) {
	var (
		sb   strings.Builder
		str  string
		tag  string
		isUp bool
	)

	switch el.kind {
	case elKindSectionL1:
		tag = `SH`
		isUp = true
	default:
		tag = `SS`
	}

	el.title.toRoff(&sb, doc)
	str = sb.String()
	str = strings.TrimSpace(str)
	str = roffEscapeString(str)
	if isUp {
		str = strings.ToUpper(str)
	}
	fmt.Fprintf(out, _roffTmplSection, tag, str)
}

func roffWriteText(out io.Writer, text []byte) {
	log.Printf(`roffWriteText: %s`, text)
	text = roffEscape(text)
	fmt.Fprint(out, string(text))
}

func roffWriteUrl(out io.Writer, doc *Document, el *element) {
	var (
		href = el.Attrs[attrNameHref]

		label strings.Builder
	)
	if el.child == nil {
		label.WriteString(el.rawStyle)
		if href == label.String() {
			label.Reset()
		}
	} else {
		el.child.toRoff(&label, doc)

		// Prevent the child being rendered later.
		el.child = nil
	}
	fmt.Fprintf(out, _roffTmplUrl, el.Attrs[attrNameHref], label.String())
}
