// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

const (
	elKindUnknown                int = iota
	elKindDocHeader                  // Wrapper.
	elKindPreamble                   // Wrapper.
	elKindDocContent                 // Wrapper.
	elKindSectionL0                  // Line started with "="
	elKindSectionL1                  // Line started with "=="
	elKindSectionL2                  // 5: Line started with "==="
	elKindSectionL3                  // Line started with "===="
	elKindSectionL4                  // Line started with "====="
	elKindSectionL5                  // Line started with "======"
	elKindSectionDiscrete            // "[discrete]"
	elKindParagraph                  // 10: Wrapper.
	elKindLiteralParagraph           // Line start with space
	elKindBlockAudio                 // "audio::"
	elKindBlockExample               // "===="
	elKindBlockExcerpts              // "____"
	elKindBlockImage                 // "image::"
	elKindBlockListing               // "----"
	elKindBlockListingNamed          // "[listing]"
	elKindBlockLiteral               // "...."
	elKindBlockLiteralNamed          // "[literal]"
	elKindBlockOpen                  // 20: Block wrapped with "--"
	elKindBlockPassthrough           // Block wrapped with "++++"
	elKindBlockSidebar               // "****"
	elKindBlockVideo                 // "video::"
	elKindCrossReference             // "<<" REF ("," LABEL) ">>"
	elKindFootnote                   // footnote:id[]
	elKindInlineID                   // "[[" REF_ID "]]" TEXT
	elKindInlineIDShort              // "[#" REF_ID "]#" TEXT "#"
	elKindInlineImage                // Inline macro for "image:"
	elKindInlinePass                 // Inline macro for passthrough "pass:"
	elKindInlineParagraph            //
	elKindListOrdered                // Wrapper.
	elKindListOrderedItem            // 30: Line start with ". "
	elKindListUnordered              // Wrapper.
	elKindListUnorderedItem          // Line start with "* " or "- "
	elKindListDescription            // Wrapper.
	elKindListDescriptionItem        // Line that has "::" + WSP
	elKindMacroTOC                   // "toc::[]"
	elKindPassthrough                // Text wrapped inside "+"
	elKindPassthroughDouble          // Text wrapped inside "++"
	elKindPassthroughTriple          // Text wrapped inside "+++"
	elKindSymbolQuoteDoubleBegin     // The ("`)
	elKindSymbolQuoteDoubleEnd       // 40: The (`")
	elKindSymbolQuoteSingleBegin     // The ('`)
	elKindSymbolQuoteSingleEnd       // The (`')
	elKindTable                      // "|==="
	elKindText                       //
	elKindTextBold                   // Text wrapped by "*"
	elKindTextItalic                 // Text wrapped by "_"
	elKindTextMono                   // Text wrapped by "`"
	elKindTextSubscript              // Word wrapped by '~'
	elKindTextSuperscript            // Word wrapped by '^'
	elKindUnconstrainedBold          // 50: Text wrapped by "**"
	elKindUnconstrainedItalic        // Text wrapped by "__"
	elKindUnconstrainedMono          // Text wrapped by "``"
	elKindURL                        // Anchor text.
	lineKindAdmonition               // "LABEL: WSP"
	lineKindAttribute                // ":" ATTR_NAME ":" (ATTR_VALUE)
	lineKindAttributeElement         // "[" ATTR_NAME ("=" ATTR_VALUE)"]"
	lineKindBlockComment             // Block start and end with "////"
	lineKindBlockTitle               // Line start with ".<alnum>"
	lineKindComment                  // Line start with "//"
	lineKindEmpty                    // 60: LF
	lineKindHorizontalRule           // "'''", "---", "- - -", "***", "* * *"
	lineKindID                       // "[[" REF_ID "]]"
	lineKindIDShort                  // "[#" REF_ID "]#" TEXT "#"
	lineKindInclude                  // "include::"
	lineKindListContinue             // "+" LF
	lineKindPageBreak                // "<<<"
	lineKindStyleClass               // "[.x.y]"
	lineKindText                     // 1*VCHAR
)

const (
	attrNameAlign       = `align`
	attrNameAlt         = `alt`
	attrNameAttribution = `attribution`
	attrNameCaption     = `caption`
	attrNameCitation    = `citation`
	attrNameCols        = `cols`
	attrNameDiscrete    = `discrete`
	attrNameEnd         = `end`
	attrNameFloat       = `float`
	attrNameFrame       = `frame`
	attrNameGrid        = `grid`
	attrNameHeight      = `height`
	attrNameHref        = `href`
	attrNameIcons       = `icons`
	attrNameLang        = `lang`
	attrNameLink        = `link`
	attrNameOptions     = `options`
	attrNameOpts        = `opts`
	attrNamePoster      = `poster`
	attrNameRefText     = `reftext`
	attrNameRel         = `rel`
	attrNameRole        = `role`
	attrNameSource      = `source`
	attrNameSrc         = `src`
	attrNameStart       = `start`
	attrNameStripes     = `stripes`
	attrNameTarget      = `target`
	attrNameTheme       = `theme`
	attrNameTitle       = `title`
	attrNameVimeo       = `vimeo`
	attrNameWidth       = `width`
	attrNameYoutube     = `youtube`
	attrNameYoutubeLang = `hl`
)

const (
	attrValueAll       = `all`
	attrValueAuthor    = `author`
	attrValueBare      = `bare`
	attrValueBlank     = `_blank`
	attrValueCols      = `cols`
	attrValueContent   = `content`
	attrValueEmail     = `email`
	attrValueEven      = `even`
	attrValueFont      = `font`
	attrValueFooter    = `footer`
	attrValueHeader    = `header`
	attrValueHover     = `hover`
	attrValueImage     = `image`
	attrValueNoopener  = `noopener`
	attrValueNoHeader  = `noheader`
	attrValueNone      = `none`
	attrValueOdd       = `odd`
	attrValueRevDate   = `revdate`
	attrValueRevNumber = `revnumber`
	attrValueRows      = `rows`
	attrValueSides     = `sides`
	attrValueTitle     = attrNameTitle
	attrValueTopbot    = `topbot`
)

const (
	classNameArabic       = `arabic`
	classNameChecklist    = `checklist`
	classNameFitContent   = `fit-content`
	classNameFrameAll     = `frame-all`
	classNameFrameEnds    = `frame-ends`
	classNameFrameNone    = `frame-none`
	classNameFrameSides   = `frame-sides`
	classNameGridAll      = `grid-all`
	classNameGridCols     = `grid-cols`
	classNameGridNone     = `grid-none`
	classNameGridRows     = `grid-rows`
	classNameLoweralpha   = `loweralpha`
	classNameLowerroman   = `lowerroman`
	classNameStretch      = `stretch`
	classNameStripesAll   = `stripes-all`
	classNameStripesEven  = `stripes-even`
	classNameStripesHover = `stripes-hover`
	classNameStripesOdd   = `stripes-odd`
	classNameTableblock   = `tableblock`
	classNameUpperalpha   = `upperalpha`
	classNameUpperroman   = `upperroman`
)

const (
	optNameAutoplay               = `autoplay`
	optNameAutowidth              = `autowidth`
	optNameControls               = `controls`
	optNameLoop                   = `loop`
	optNameNocontrols             = `nocontrols`
	optVideoFullscreen            = `fs`
	optVideoModest                = `modest`
	optVideoNofullscreen          = `nofullscreen`
	optVideoPlaylist              = `playlist`
	optVideoYoutubeModestbranding = `modestbranding`
)

const (
	prefixInclude = `include::`
)

const (
	admonitionCaution   = `CAUTION`
	admonitionImportant = `IMPORTANT`
	admonitionNote      = `NOTE`
	admonitionTip       = `TIP`
	admonitionWarning   = `WARNING`
)

const (
	_                    int64 = iota
	styleSectionColophon       = 1 << (iota - 1)
	styleSectionAbstract
	styleSectionDiscrete
	styleSectionPreface
	styleSectionDedication
	styleSectionPartIntroduction
	styleSectionAppendix
	styleSectionGlossary
	styleSectionBibliography
	styleSectionIndex
	styleListMarkerCircle
	styleListMarkerDisc
	styleListMarkerNone
	styleListMarkerSquare
	styleListMarkerUnstyled
	styleParagraphLead
	styleParagraphNormal
	styleLink
	styleNumberingArabic
	styleNumberingDecimal
	styleNumberingLoweralpha
	styleNumberingUpperalpha
	styleNumberingLowerroman
	styleNumberingUpperroman
	styleNumberingLowergreek
	styleDescriptionHorizontal
	styleDescriptionQandA
	styleAdmonition
	styleBlockListing
	styleQuote
	styleSource
	styleTextBold
	styleTextItalic
	styleTextMono
	styleVerse
)

const (
	symbolQuoteDoubleBegin = `&#8220;`
	symbolQuoteDoubleEnd   = `&#8221;`
	symbolQuoteSingleBegin = `&#8216;`
	symbolQuoteSingleEnd   = `&#8217;`
	symbolChecked          = `&#10003;`
	symbolUnchecked        = `&#10063;`
)

var adocStyles = map[string]int64{
	`colophon`:          styleSectionColophon,
	`abstract`:          styleSectionAbstract,
	`preface`:           styleSectionPreface,
	`dedication`:        styleSectionDedication,
	attrNameDiscrete:    styleSectionDiscrete,
	`partintro`:         styleSectionPartIntroduction,
	`appendix`:          styleSectionAppendix,
	`glossary`:          styleSectionGlossary,
	`bibliography`:      styleSectionBibliography,
	`index`:             styleSectionIndex,
	`circle`:            styleListMarkerCircle,
	`disc`:              styleListMarkerDisc,
	`none`:              styleListMarkerNone,
	`square`:            styleListMarkerSquare,
	`unstyled`:          styleListMarkerUnstyled,
	`.lead`:             styleParagraphLead,
	`.normal`:           styleParagraphNormal,
	`arabic`:            styleNumberingArabic,
	`decimal`:           styleNumberingDecimal,
	`loweralpha`:        styleNumberingLoweralpha,
	`upperalpha`:        styleNumberingUpperalpha,
	`lowerroman`:        styleNumberingLowerroman,
	`upperroman`:        styleNumberingUpperroman,
	`lowergreek`:        styleNumberingLowergreek,
	`horizontal`:        styleDescriptionHorizontal,
	`qanda`:             styleDescriptionQandA,
	admonitionCaution:   styleAdmonition,
	admonitionImportant: styleAdmonition,
	admonitionNote:      styleAdmonition,
	admonitionTip:       styleAdmonition,
	admonitionWarning:   styleAdmonition,
	`listing`:           styleBlockListing,
	`quote`:             styleQuote,
	`source`:            styleSource,
	`verse`:             styleVerse,
}

var _attrRef = map[string]string{
	`amp`:            `&`,
	`apos`:           htmlSymbolSingleQuote, // '
	`asterisk`:       `*`,
	`backslash`:      `\`,
	`backtick`:       "`",
	`blank`:          ``,
	`brvbar`:         htmlSymbolBrokenVerticalBar, // ¦
	`caret`:          `^`,
	`cpp`:            `C++`,
	`deg`:            htmlSymbolDegreeSign, // °
	`empty`:          ``,
	`endsb`:          `]`,
	`gt`:             `>`,
	`ldquo`:          htmlSymbolLeftDoubleQuote,
	`lsquo`:          htmlSymbolLeftSingleQuote,
	`lt`:             `<`,
	`nbsp`:           htmlSymbolNonBreakingSpace,
	`plus`:           htmlSymbolPlus,
	`quot`:           htmlSymbolDoubleQuote,
	`rdquo`:          htmlSymbolRightDoubleQuote,
	`rsquo`:          htmlSymbolRightSingleQuote,
	`sp`:             ` `,
	`startsb`:        `[`,
	`tilde`:          `~`,
	`two-colons`:     `::`,
	`two-semicolons`: `;;`,
	`vbar`:           `|`,
	`wj`:             htmlSymbolWordJoiner,
	`zwsp`:           htmlSymbolZeroWidthSpace,
}

// preprocessBlockCode preprocess the content of block code, like "include::"
// directive, and return the new content.
func preprocessBlockCode(doc *Document, content []byte) (newContent []byte) {
	var bbuf bytes.Buffer
	var lines = bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte(`include::`)) {
			var elInclude = parseInclude(doc, line)
			if elInclude != nil {
				bbuf.Write(elInclude.content)
				bbuf.WriteByte('\n')
				continue
			}
		}
		bbuf.Write(line)
		bbuf.WriteByte('\n')
	}

	newContent = applySubstitutions(doc, bbuf.Bytes())
	return newContent
}

// applySubstitutions scan the content and replace attribute reference "{}"
// with its value, and character '<', '>', '&' with HTML symbol.
func applySubstitutions(doc *Document, content []byte) []byte {
	var (
		raw    = bytes.TrimRight(content, " \n")
		newraw = make([]byte, 0, len(raw))
		buf    = bytes.NewBuffer(newraw)

		c      byte
		x      int
		newRaw []byte
		ok     bool
	)
	for x < len(raw) {
		c = raw[x]

		if c == '{' {
			newRaw, ok = parseAttrRef(doc, raw, x)
			if ok {
				raw = newRaw
				continue
			}
			buf.WriteByte(c)
		} else if c == '<' {
			buf.WriteString(htmlSymbolLessthan)
		} else if c == '>' {
			buf.WriteString(htmlSymbolGreaterthan)
		} else if c == '&' {
			buf.WriteString(htmlSymbolAmpersand)
		} else {
			buf.WriteByte(c)
		}
		x++
	}
	return buf.Bytes()
}

// generateID generate ID for anchor.
// This function follow the [Mozilla specification].
//
// The generated ID is affected by the following document attributes:
// `idprefix` and `idseparator`.
//
// The idprefix must be ASCII string.
// It must start with '_', '-', or ASCII letters, otherwise the '_' will be
// prepended.
// If one of the character is not valid, it will replaced with '_'.
//
// The `idseparator` can be empty or single ASCII character ('_' or '-',
// ASCII letter, or digit).
// It is used to replace invalid characters in the src.
//
// [Mozilla specification]: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id.
func generateID(doc *Document, str string) string {
	var (
		idSep byte = '_'

		v    string
		bout []byte
		c    byte
		ok   bool
	)

	v, ok = doc.Attributes.Entry[docAttrIDPrefix]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			str = v + str
		}
	}

	bout = make([]byte, 0, len(str))

	v, ok = doc.Attributes.Entry[docAttrIDSeparator]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			// idseparator document attribute exist and set to
			// empty.
			idSep = 0
		} else {
			c = v[0]
			if c == '_' || c == '-' || ascii.IsAlnum(c) {
				idSep = c
			}
		}
	}

	for _, c = range []byte(str) {
		if c == '_' || c == '-' || ascii.IsAlnum(c) {
			if c >= 'A' && c <= 'Z' {
				bout = append(bout, c+32)
			} else {
				bout = append(bout, c)
			}
		} else if idSep != 0 {
			bout = append(bout, idSep)
		}
	}

	if len(bout) == 0 {
		bout = append(bout, '_')
	} else if !ascii.IsAlpha(bout[0]) && bout[0] != '_' {
		bout = append(bout, '_')
		copy(bout[1:], bout)
		bout[0] = '_'
	}

	return string(bout)
}

func isAdmonition(line []byte) bool {
	var x int
	switch {
	case bytes.HasPrefix(line, []byte(admonitionCaution)):
		x = len(admonitionCaution)
	case bytes.HasPrefix(line, []byte(admonitionImportant)):
		x = len(admonitionImportant)
	case bytes.HasPrefix(line, []byte(admonitionNote)):
		x = len(admonitionNote)
	case bytes.HasPrefix(line, []byte(admonitionTip)):
		x = len(admonitionTip)
	case bytes.HasPrefix(line, []byte(admonitionWarning)):
		x = len(admonitionWarning)
	default:
		return false
	}
	if x >= len(line) {
		return false
	}
	if line[x] == ':' {
		x++
		if x >= len(line) {
			return false
		}
		if line[x] == ' ' || line[x] == '\t' {
			return true
		}
	}
	return false
}

func isLineDescriptionItem(line []byte) bool {
	var (
		x int
	)

	_, x = indexUnescape(line, []byte(`:: `))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(line, []byte("::\t"))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(line, []byte(`::`))
	return x > 0
}

func isStyleAdmonition(style int64) bool {
	return style&styleAdmonition > 0
}

func isStyleQuote(style int64) bool {
	return style&styleQuote > 0
}

func isStyleVerse(style int64) bool {
	return style&styleVerse > 0
}

// isValidID will return true if id is valid HTML ref ID according to
// [Mozilla specification], where the first character is either '-', '_', or
// ASCII letter, and the rest is either '-', '_', ASCII letter or digit.
//
// [Mozilla specification]: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id.
func isValidID(id []byte) bool {
	var (
		x int
		c byte
	)

	for x, c = range id {
		if x == 0 {
			if !(c == '-' || c == '_' || ascii.IsAlpha(c)) {
				return false
			}
			continue
		}
		if c == '-' || c == '_' || c == '.' || ascii.IsAlnum(c) {
			continue
		}
		return false
	}
	return true
}

// parseAttrRef parse the attribute reference, an attribute key wrapped by
// "{" "}".  If the attribute reference exist, replace the content with the
// attribute value and reset the parser state to zero.
func parseAttrRef(doc *Document, content []byte, x int) (newContent []byte, ok bool) {
	var (
		raw = content[x+1:]

		name      string
		attrValue string
		attrName  []byte
		rest      []byte
		idx       int
	)

	attrName, idx = indexByteUnescape(raw, '}')
	if idx < 0 {
		return nil, false
	}

	name = string(bytes.TrimSpace(bytes.ToLower(attrName)))
	attrValue, ok = _attrRef[name]
	if !ok {
		attrValue, ok = doc.Attributes.Entry[name]
		if !ok {
			return nil, false
		}

		// Add prefix "mailto:" if the ref name start with email, so
		// it can be parsed by caller as macro link.
		if name == `email` || strings.HasPrefix(name, `email_`) {
			attrValue = `mailto:` + attrValue + `[` + attrValue + `]`
		}
	}

	rest = content[x+idx+2:]
	newContent = make([]byte, 0, len(attrValue)+len(rest))
	newContent = append(newContent, attrValue...)
	newContent = append(newContent, rest...)
	return newContent, true
}

// parseClosedBracket parse the text in input until we found the last close
// bracket.
// It will skip any open-close brackets inside input.
// For example, parsing ("test:[]]", '[', ']') will return ("test:[]", 7).
//
// If no closed bracket found it will return (nil, -1).
func parseClosedBracket(input []byte, openb, closedb byte) (out []byte, idx int) {
	var (
		openCount int
		c         byte
		isEsc     bool
	)

	out = make([]byte, 0, len(input))

	for idx, c = range input {
		if c == '\\' {
			if isEsc {
				out = append(out, '\\')
				isEsc = false
			} else {
				isEsc = true
			}
			continue
		}

		if c == closedb {
			if isEsc {
				out = append(out, c)
				isEsc = false
				continue
			}
			if openCount == 0 {
				return out, idx
			}
			openCount--
			out = append(out, c)
			continue
		}

		if c == openb {
			out = append(out, c)
			if isEsc {
				isEsc = false
			} else {
				openCount++
			}
			continue
		}

		if isEsc {
			out = append(out, '\\')
			isEsc = false
		}
		out = append(out, c)
	}

	// No closed bracket found.
	return nil, -1
}

// parseIDLabel parse the string "ID (,LABEL)" into id and label.
// It will return empty id and label if ID is not valid.
func parseIDLabel(s []byte) (id, label []byte) {
	var idLabel = bytes.Split(s, []byte(`,`))

	id = idLabel[0]
	if len(idLabel) >= 2 {
		label = idLabel[1]
	}
	if isValidID(idLabel[0]) {
		return id, label
	}
	return nil, nil
}

func parseInlineMarkup(doc *Document, content []byte) (container *element) {
	var pi = newInlineParser(doc, content)

	pi.do()
	return pi.container
}

// parseStyle get the style based on string value.
func parseStyle(styleName string) (styleKind int64) {
	// Check for admonition label first...
	styleKind = adocStyles[styleName]
	if styleKind > 0 {
		return styleKind
	}

	styleName = strings.ToLower(styleName)
	styleKind = adocStyles[styleName]
	if styleKind > 0 {
		return styleKind
	}

	return 0
}
