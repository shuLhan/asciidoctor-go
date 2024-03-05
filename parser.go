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

// List of document metadata.
const (
	MetaNameAuthor      = `author`       // May contain the first author full name only.
	MetaNameAuthorNames = `author_names` // List of author full names, separated by comma.
	MetaNameDescription = `description`
	MetaNameGenerator   = `generator`
	MetaNameKeywords    = `keywords`

	metaNameAuthorInitials  = `authorinitials`
	metaNameDocTitle        = `doctitle`
	metaNameEmail           = attrValueEmail
	metaNameFirstName       = `firstname`
	metaNameIDPrefix        = `idprefix`
	metaNameIDSeparator     = `idseparator`
	metaNameLastName        = `lastname`
	metaNameLastUpdateLabel = `last-update-label`
	metaNameLastUpdateValue = `last-update-value`
	metaNameMiddleName      = `middlename`
	metaNameNoFooter        = `nofooter`
	metaNameNoHeader        = `noheader`
	metaNameNoHeaderFooter  = `no-header-footer`
	metaNameNoTitle         = `notitle`
	metaNameRevDate         = `revdate`
	metaNameRevNumber       = `revnumber`
	metaNameRevRemark       = `revremark`
	metaNameSectAnchors     = `sectanchors`
	metaNameSectIDs         = `sectids`
	metaNameSectLinks       = `sectlinks`
	metaNameSectNumLevel    = `sectnumlevels`
	metaNameSectNums        = `sectnums`
	metaNameShowTitle       = `showtitle`
	metaNameTOC             = `toc`
	metaNameTOCLevels       = `toclevels`
	metaNameTOCTitle        = `toc-title`
	metaNameTableCaption    = `table-caption`
	metaNameTitle           = attrNameTitle
	metaNameTitleSeparator  = `title-separator`
	metaNameVersionLabel    = `version-label`
)

// List of possible metadata value.
const (
	metaValueAuto     = `auto`
	metaValueMacro    = `macro`
	metaValuePreamble = `preamble`
	metaValueLeft     = `left`
	metaValueRight    = `right`
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

		// We use if-condition with "continue" to break and continue
		// the for-loop, so it is not possible to use switch-case
		// here.
		//
		//nolint:gocritic
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
// The generated ID is affected by the following metadata: `idprefix` and
// `idseparator`.
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

	v, ok = doc.Attributes[metaNameIDPrefix]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			str = v + str
		}
	}

	bout = make([]byte, 0, len(str))

	v, ok = doc.Attributes[metaNameIDSeparator]
	if ok {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			// idseparator metadata exist and set to empty.
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

func isTitle(line []byte) bool {
	if line[0] == '=' || line[0] == '#' {
		if len(line) > 1 && (line[1] == ' ' || line[1] == '\t') {
			return true
		}
	}
	return false
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
		attrValue, ok = doc.Attributes[name]
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

// whatKindOfLine return the kind of line.
// It will return lineKindText if the line does not match with known syntax.
func whatKindOfLine(line []byte) (kind int, spaces, got []byte) {
	kind = lineKindText

	line = bytes.TrimRight(line, " \f\n\r\t\v")

	// All of the comparison MUST be in order.

	if len(line) == 0 {
		return lineKindEmpty, nil, line
	}
	if bytes.HasPrefix(line, []byte(`////`)) {
		// Check for comment block first, since we use HasPrefix to
		// check for single line comment.
		return lineKindBlockComment, spaces, line
	}
	if bytes.HasPrefix(line, []byte(`//`)) {
		// Use HasPrefix to allow single line comment without space,
		// for example "//comment".
		return lineKindComment, spaces, line
	}

	var strline = string(line)

	switch strline {
	case `'''`, `---`, `- - -`, `***`, `* * *`:
		return lineKindHorizontalRule, spaces, line
	case `<<<`:
		return lineKindPageBreak, spaces, line
	case `--`:
		return elKindBlockOpen, spaces, line
	case `____`:
		return elKindBlockExcerpts, spaces, line
	case `....`:
		return elKindBlockLiteral, nil, line
	case `++++`:
		return elKindBlockPassthrough, spaces, line
	case `****`:
		return elKindBlockSidebar, nil, line
	case `====`:
		return elKindBlockExample, spaces, line
	case `[listing]`:
		return elKindBlockListingNamed, nil, line
	case `[literal]`:
		return elKindBlockLiteralNamed, nil, line
	case `toc::[]`:
		return elKindMacroTOC, spaces, line
	}

	if bytes.HasPrefix(line, []byte(`|===`)) {
		return elKindTable, nil, line
	}
	if bytes.HasPrefix(line, []byte(`image::`)) {
		return elKindBlockImage, spaces, line
	}
	if bytes.HasPrefix(line, []byte(`include::`)) {
		return lineKindInclude, nil, line
	}
	if bytes.HasPrefix(line, []byte(`video::`)) {
		return elKindBlockVideo, nil, line
	}
	if bytes.HasPrefix(line, []byte(`audio::`)) {
		return elKindBlockAudio, nil, line
	}
	if isAdmonition(line) {
		return lineKindAdmonition, nil, line
	}

	var (
		x        int
		r        byte
		hasSpace bool
	)
	for x, r = range line {
		if r == ' ' || r == '\t' {
			hasSpace = true
			continue
		}
		break
	}
	if hasSpace {
		spaces = line[:x]
		line = line[x:]

		// A line indented with space only allowed on list item,
		// otherwise it would be set as literal paragraph.

		if isLineDescriptionItem(line) {
			return elKindListDescriptionItem, spaces, line
		}

		if line[0] != '*' && line[0] != '-' && line[0] != '.' {
			return elKindLiteralParagraph, spaces, line
		}
	}

	switch line[0] {
	case ':':
		kind = lineKindAttribute
	case '[':
		var (
			newline = bytes.TrimRight(line, " \t")
			l       = len(newline)
		)

		if newline[l-1] != ']' {
			return lineKindText, nil, line
		}
		if l >= 5 {
			// [[x]]
			if newline[1] == '[' && newline[l-2] == ']' {
				return lineKindID, nil, line
			}
		}
		if l >= 4 {
			// [#x]
			if line[1] == '#' {
				return lineKindIDShort, nil, line
			}
			// [.x]
			if line[1] == '.' {
				return lineKindStyleClass, nil, line
			}
		}
		return lineKindAttributeElement, spaces, line
	case '=':
		var subs = bytes.Fields(line)

		switch string(subs[0]) {
		case `==`:
			kind = elKindSectionL1
		case `===`:
			kind = elKindSectionL2
		case `====`:
			kind = elKindSectionL3
		case `=====`:
			kind = elKindSectionL4
		case `======`:
			kind = elKindSectionL5
		}
	case '.':
		switch {
		case len(line) <= 1:
			kind = lineKindText
		case ascii.IsAlnum(line[1]):
			kind = lineKindBlockTitle
		default:
			x = 0
			for ; x < len(line); x++ {
				if line[x] == '.' {
					continue
				}
				if line[x] == ' ' || line[x] == '\t' {
					kind = elKindListOrderedItem
					return kind, spaces, line
				}
			}
		}
	case '*', '-':
		if len(line) <= 1 {
			kind = lineKindText
			return kind, spaces, line
		}

		var (
			listItemChar = line[0]
			count        = 0
		)
		x = 0
		for ; x < len(line); x++ {
			if line[x] == listItemChar {
				count++
				continue
			}
			if line[x] == ' ' || line[x] == '\t' {
				kind = elKindListUnorderedItem
				return kind, spaces, line
			}
			// Break on the first non-space, so from above
			// condition we have,
			// - item
			// -- item
			// --- item
			// ---- // block listing
			// --unknown // break here
			break
		}
		if listItemChar == '-' && count == 4 && x == len(line) {
			kind = elKindBlockListing
		} else {
			kind = lineKindText
		}
		return kind, spaces, line
	default:
		switch string(line) {
		case `+`:
			kind = lineKindListContinue
		case `----`:
			kind = elKindBlockListing
		default:
			if isLineDescriptionItem(line) {
				kind = elKindListDescriptionItem
			}
		}
	}
	return kind, spaces, line
}
