// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	libascii "git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
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

// htmlSubs apply the text substitutions to element.raw based on applySubs in
// the following order: c, q, a, r, m, p.
// If applySubs is 0, it will return element.raw as is.
func htmlSubs(doc *Document, el *element) []byte {
	var (
		input = el.raw
	)
	if el.applySubs == 0 {
		return input
	}
	if el.applySubs&passSubChar != 0 {
		input = htmlSubsChar(input)
	}
	if el.applySubs&passSubQuote != 0 {
		input = htmlSubsQuote(input)
	}
	if el.applySubs&passSubAttr != 0 {
		input = htmlSubsAttr(doc, input)
	}
	if el.applySubs&passSubRepl != 0 {
		input = htmlSubsRepl(input)
	}
	if el.applySubs&passSubMacro != 0 {
		input = htmlSubsMacro(doc, input, el.kind == elKindInlinePass)
	}
	return input
}

// htmlSubsChar replace character '<', '>', and '&' with "&lt;", "&gt;", and
// "&amp;".
//
// Ref: https://docs.asciidoctor.org/asciidoc/latest/subs/special-characters/
func htmlSubsChar(input []byte) []byte {
	var (
		bb bytes.Buffer
		c  byte
	)
	for _, c = range input {
		if c == '<' {
			bb.WriteString(`&lt;`)
			continue
		}
		if c == '>' {
			bb.WriteString(`&gt;`)
			continue
		}
		if c == '&' {
			bb.WriteString(`&amp;`)
			continue
		}
		bb.WriteByte(c)
	}
	return bb.Bytes()
}

// htmlSubsQuote replace inline markup with its HTML markup.
// The following inline markup ara parsed and substitutes,
//
//   - emphasis: _word_ with "<em>word</em>".
//   - strong: *word* with "<strong>word</strong>".
//   - monospace: `word` with "<code>word</code>".
//   - superscript: ^word^ with "<sup>word</sup>".
//   - subscript: ~word~ with "<sub>word</sub>".
//   - double curved quotes: "`word`" with "&#8220;word&#8221;"
//   - single curved quotes: '`word`' with "&#8216;word&#8217;"
//
// Ref: https://docs.asciidoctor.org/asciidoc/latest/subs/quotes/
func htmlSubsQuote(input []byte) []byte {
	var (
		bb    bytes.Buffer
		x     int
		idx   int
		text  []byte
		c1    byte
		nextc byte
	)
	for x < len(input) {
		c1 = input[x]

		x++
		if x == len(input) {
			// Nothing left to parsed.
			bb.WriteByte(c1)
			break
		}
		nextc = input[x]

		if c1 == '_' {
			text, idx = indexByteUnescape(input[x:], c1)
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(`<em>`)
			bb.Write(text)
			bb.WriteString(`</em>`)
			x = x + idx + 1
			continue
		}
		if c1 == '*' {
			text, idx = indexByteUnescape(input[x:], c1)
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(`<strong>`)
			bb.Write(text)
			bb.WriteString(`</strong>`)
			x = x + idx + 1
			continue
		}
		if c1 == '`' {
			text, idx = indexByteUnescape(input[x:], c1)
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(`<code>`)
			bb.Write(text)
			bb.WriteString(`</code>`)
			x = x + idx + 1
			continue
		}
		if c1 == '^' {
			text, idx = indexByteUnescape(input[x:], c1)
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(`<sup>`)
			bb.Write(text)
			bb.WriteString(`</sup>`)
			x = x + idx + 1
			continue
		}
		if c1 == '~' {
			text, idx = indexByteUnescape(input[x:], c1)
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(`<sub>`)
			bb.Write(text)
			bb.WriteString(`</sub>`)
			x = x + idx + 1
			continue
		}
		if c1 == '"' {
			if nextc != '`' {
				bb.WriteByte(c1)
				continue
			}
			if x+1 == len(input) {
				bb.WriteByte(c1)
				continue
			}

			text, idx = indexUnescape(input[x+1:], []byte("`\""))
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(htmlSymbolLeftDoubleQuote)
			bb.Write(text)
			bb.WriteString(htmlSymbolRightDoubleQuote)
			x = x + idx + 3
			continue
		}
		if c1 == '\'' {
			if nextc != '`' {
				bb.WriteByte(c1)
				continue
			}
			if x+1 == len(input) {
				bb.WriteByte(c1)
				continue
			}

			text, idx = indexUnescape(input[x+1:], []byte("`'"))
			if text == nil {
				bb.WriteByte(c1)
				continue
			}
			bb.WriteString(htmlSymbolLeftSingleQuote)
			bb.Write(text)
			bb.WriteString(htmlSymbolRightSingleQuote)
			x = x + idx + 3
			continue
		}
		bb.WriteByte(c1)
	}
	return bb.Bytes()
}

// htmlSubsAttr replace attribute (the `{...}`) with its values.
//
// Ref: https://docs.asciidoctor.org/asciidoc/latest/subs/attributes/
func htmlSubsAttr(doc *Document, input []byte) []byte {
	var (
		bb     bytes.Buffer
		key    string
		val    string
		vbytes []byte
		idx    int
		x      int
		c      byte
		ok     bool
	)

	for x < len(input) {
		c = input[x]
		x++
		if c != '{' {
			bb.WriteByte(c)
			continue
		}

		vbytes, idx = indexByteUnescape(input[x:], '}')
		if vbytes == nil {
			bb.WriteByte(c)
			continue
		}
		vbytes = bytes.TrimSpace(vbytes)
		vbytes = bytes.ToLower(vbytes)

		key = string(vbytes)
		val, ok = _attrRef[key]
		if ok {
			bb.WriteString(val)
			x = x + idx + 1
			continue
		}

		val, ok = doc.Attributes[key]
		if !ok {
			bb.WriteByte(c)
			continue
		}

		// Add prefix "mailto:" if the ref name start with email, so
		// it can be parsed by caller as macro link.
		if key == `email` || strings.HasPrefix(key, `email_`) {
			val = `mailto:` + val + `[` + val + `]`
		}

		bb.WriteString(val)
		x = x + idx + 1
	}

	return bb.Bytes()
}

// htmlSubsRepl substitutes special characters with HTML unicode.
//
// The special characters are,
//
//   - (C) replaced with &#169;
//   - (R)  : &#174;
//   - (TM) : &#8482;
//   - --   : &#8212; Only replaced if between two word characters, between a
//     word character and a line boundary, or flanked by spaces.
//     When flanked by space characters (e.g., a -- b), the normal spaces are
//     replaced by thin spaces (&#8201;).
//   - ...  : &#8230;
//   - ->   : &#8594;
//   - =>   : &#8658;
//   - <-   : &#8592;
//   - <=   : &#8656;
//   - '    : &#8217;
//
// According to [the documentation], this substitution step also recognizes
// HTML and XML character references as well as decimal and hexadecimal
// Unicode code points, but we only cover the above right now.
//
// [the documentation]: https://docs.asciidoctor.org/asciidoc/latest/subs/replacements/
func htmlSubsRepl(input []byte) (out []byte) {
	var (
		text  []byte
		x     int
		idx   int
		c1    byte
		nextc byte
		prevc byte
	)

	out = make([]byte, 0, len(input))

	for x < len(input) {
		prevc = c1
		c1 = input[x]

		x++
		if x == len(input) {
			out = append(out, c1)
			break
		}
		nextc = input[x]

		if c1 == '(' {
			text, idx = indexByteUnescape(input[x:], ')')
			if len(text) == 1 {
				if text[0] == 'C' {
					out = append(out, []byte(htmlSymbolCopyright)...)
					x = x + idx + 1
					c1 = ')'
					continue
				}
				if text[0] == 'R' {
					out = append(out, []byte(htmlSymbolRegistered)...)
					x = x + idx + 1
					c1 = ')'
					continue
				}
			} else if len(text) == 2 {
				if text[0] == 'T' && text[1] == 'M' {
					out = append(out, []byte(htmlSymbolTrademark)...)
					x = x + idx + 1
					c1 = ')'
					continue
				}
			}

			out = append(out, c1)
			continue
		}
		if c1 == '-' {
			if nextc == '>' {
				out = append(out, []byte(htmlSymbolSingleRightArrow)...)
				x++
				c1 = nextc
				continue
			}
			if nextc == '-' {
				if x+1 >= len(input) {
					out = append(out, c1)
					continue
				}
				// set c1 to the third character after '--'.
				c1 = input[x+1]
				if libascii.IsSpace(prevc) && libascii.IsSpace(c1) {
					out = out[:len(out)-1]
					out = append(out, []byte(htmlSymbolThinSpace)...)
					out = append(out, []byte(htmlSymbolEmdash)...)
					out = append(out, []byte(htmlSymbolThinSpace)...)
					x += 2
					continue
				}
				if libascii.IsAlpha(prevc) && libascii.IsAlpha(c1) {
					out = append(out, []byte(htmlSymbolEmdash)...)
					x++
					continue
				}
			}
			out = append(out, c1)
			continue
		}
		if c1 == '=' {
			if nextc == '>' {
				out = append(out, []byte(htmlSymbolDoubleRightArrow)...)
				x++
				c1 = nextc
				continue
			}
			out = append(out, c1)
			continue
		}
		if c1 == '<' {
			if nextc == '-' {
				out = append(out, []byte(htmlSymbolSingleLeftArrow)...)
				x++
				continue
			}
			if nextc == '=' {
				out = append(out, []byte(htmlSymbolDoubleLeftArrow)...)
				x++
				continue
			}
			out = append(out, c1)
			continue
		}
		if c1 == '.' {
			if nextc != '.' {
				out = append(out, c1)
				continue
			}
			if x+1 >= len(input) {
				out = append(out, c1)
				continue
			}
			// Set c1 to the third character.
			c1 = input[x+1]
			if c1 == '.' {
				out = append(out, []byte(htmlSymbolEllipsis)...)
				x += 2
				continue
			}
			out = append(out, c1)
			continue
		}
		if c1 == '\'' {
			if libascii.IsAlpha(prevc) {
				out = append(out, []byte(htmlSymbolApostrophe)...)
				continue
			}
			out = append(out, c1)
			continue
		}
		out = append(out, c1)
	}
	return out
}

// htmlSubsMacro substitutes macro with its HTML markup.
func htmlSubsMacro(doc *Document, input []byte, isInlinePass bool) (out []byte) {
	var (
		el        *element
		bb        bytes.Buffer
		macroName string
		x         int
		n         int
		c         byte
	)

	for x < len(input) {
		c = input[x]
		if c != ':' {
			out = append(out, c)
			x++
			continue
		}

		macroName = parseMacroName(input[:x])
		if len(macroName) == 0 {
			out = append(out, c)
			x++
			continue
		}

		switch macroName {
		case macroFootnote:
			el, n = parseMacroFootnote(doc, input[x+1:])
			if el == nil {
				out = append(out, c)
				x++
				continue
			}
			x += n
			n = len(out)
			out = out[:n-len(macroName)] // Undo the macro name
			bb.Reset()
			htmlWriteFootnote(el, &bb)
			out = append(out, bb.Bytes()...)

		case macroFTP, macroHTTPS, macroHTTP, macroIRC, macroLink, macroMailto:
			el, n = parseURL(doc, macroName, input[x+1:])
			if el == nil {
				out = append(out, c)
				x++
				continue
			}
			x += n
			n = len(out)
			out = out[:n-len(macroName)]
			bb.Reset()
			htmlWriteURLBegin(el, &bb)
			if el.child != nil {
				el.child.toHTML(doc, &bb)
			}
			htmlWriteURLEnd(&bb)
			out = append(out, bb.Bytes()...)

		case macroImage:
			el, n = parseInlineImage(doc, input[x+1:])
			if el == nil {
				out = append(out, c)
				x++
				continue
			}
			x += n
			n = len(out)
			out = out[:n-len(macroName)]
			bb.Reset()
			htmlWriteInlineImage(el, &bb)
			out = append(out, bb.Bytes()...)

		case macroPass:
			if isInlinePass {
				// Prevent recursive substitutions.
				out = append(out, c)
				x++
				continue
			}
			el, n = parseMacroPass(input[x+1:])
			if el == nil {
				out = append(out, c)
				x++
				continue
			}
			x += n
			n = len(out)
			out = out[:n-len(macroName)]
			bb.Reset()
			htmlWriteInlinePass(doc, el, &bb)
			out = append(out, bb.Bytes()...)

		default:
			out = append(out, c)
			x++
		}
	}
	return out
}

func htmlWriteBlockBegin(el *element, out io.Writer, addClass string) {
	fmt.Fprint(out, "\n<div")

	if len(el.ID) > 0 {
		fmt.Fprintf(out, ` id="%s"`, el.ID)
	}

	var (
		classes = el.htmlClasses()
		c       = strings.TrimSpace(addClass + ` ` + classes)
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

	var iconsFont = el.Attrs[attrNameIcons]

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
		optControls = ` controls`
		src         = el.Attrs[attrNameSrc]

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
		v               string
		withAttribution bool
		withCitation    bool
	)

	v, withAttribution = el.Attrs[attrNameAttribution]
	if withAttribution {
		fmt.Fprintf(out, "\n<div class=%q>\n&#8212; %s", attrNameAttribution, v)
	}

	v, withCitation = el.Attrs[attrNameCitation]
	if withCitation {
		fmt.Fprintf(out, "<br>\n<cite>%s</cite>", v)
	}

	if withAttribution {
		fmt.Fprint(out, "\n</div>")
	}
	fmt.Fprint(out, "\n</div>")
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

		if doc.preamble != nil {
			fmt.Fprint(out, _lf+`<div id="preamble">`)
			fmt.Fprint(out, _lf+`<div class="sectionbody">`)

			if doc.preamble.child != nil {
				doc.preamble.child.toHTML(doc, out)
			}

			fmt.Fprint(out, _lf+`</div>`)
			if doc.tocIsEnabled && doc.tocPosition == metaValuePreamble {
				doc.tocHTML(out)
			}
			fmt.Fprint(out, _lf+`</div>`)
		}
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
	var (
		label string
		value string
		ok    bool
	)

	fmt.Fprint(out, `
<div id="footer">
<div id="footer-text">`)

	if len(doc.Revision.Number) > 0 {
		label, ok = doc.Attributes[metaNameVersionLabel]
		if ok && len(label) == 0 {
			label = `Version `
		} else {
			label = ` `
		}

		fmt.Fprintf(out, "\n%s%s<br>", label, doc.Revision.Number)
	}

	label, ok = doc.Attributes[metaNameLastUpdateLabel]
	if ok {
		value = doc.Attributes[metaNameLastUpdateValue]
		if len(value) != 0 {
			fmt.Fprintf(out, "\n%s %s", label, value)
		}
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
	fmt.Fprint(out, `<div id="header">`)

	var (
		haveHeader = doc.haveHeader()

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

	if haveHeader {
		fmt.Fprint(out, "\n<div class=\"details\">")
	}

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
	if haveHeader {
		fmt.Fprint(out, "\n</div>")
	}

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
		src = el.Attrs[attrNameSrc]
		alt = el.Attrs[attrNameAlt]

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

func htmlWriteInlinePass(doc *Document, el *element, out io.Writer) {
	var text = htmlSubs(doc, el)

	fmt.Fprint(out, string(text))
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
		class = el.getListOrderedClass()
		tipe  = el.getListOrderedType()
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
		table = el.table

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
		rows = table.rows
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
		classes = el.htmlClasses()
		target  = el.Attrs[attrNameTarget]
		rel     = el.Attrs[attrNameRel]
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
