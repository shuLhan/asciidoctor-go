// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/shuLhan/share/lib/ascii"
)

const (
	macroFTP    = "ftp"
	macroHTTP   = "http"
	macroHTTPS  = "https"
	macroIRC    = "irc"
	macroImage  = "image"
	macroLink   = "link"
	macroMailto = "mailto"
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
	elKindInlineID                   // "[[" REF_ID "]]" TEXT
	elKindInlineIDShort              // "[#" REF_ID "]#" TEXT "#"
	elKindInlineImage                // Inline macro for "image:"
	elKindInlineParagraph            //
	elKindListOrdered                // Wrapper.
	elKindListOrderedItem            // 30: Line start with ". "
	elKindListUnordered              // Wrapper.
	elKindListUnorderedItem          // Line start with "* "
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
	attrNameAlign       = "align"
	attrNameAlt         = "alt"
	attrNameAttribution = "attribution"
	attrNameCaption     = "caption"
	attrNameCitation    = "citation"
	attrNameCols        = "cols"
	attrNameDiscrete    = "discrete"
	attrNameEnd         = "end"
	attrNameFloat       = "float"
	attrNameFrame       = "frame"
	attrNameGrid        = "grid"
	attrNameHeight      = "height"
	attrNameHref        = "href"
	attrNameIcons       = "icons"
	attrNameLang        = "lang"
	attrNameLink        = "link"
	attrNameOptions     = "options"
	attrNameOpts        = "opts"
	attrNamePoster      = "poster"
	attrNameRefText     = "reftext"
	attrNameRel         = "rel"
	attrNameRole        = "role"
	attrNameSource      = "source"
	attrNameSrc         = "src"
	attrNameStart       = "start"
	attrNameStripes     = "stripes"
	attrNameTarget      = "target"
	attrNameTheme       = "theme"
	attrNameTitle       = "title"
	attrNameVimeo       = "vimeo"
	attrNameWidth       = "width"
	attrNameYoutube     = "youtube"
	attrNameYoutubeLang = "hl"
)

const (
	attrValueAll       = "all"
	attrValueAuthor    = "author"
	attrValueBare      = "bare"
	attrValueBlank     = "_blank"
	attrValueCols      = "cols"
	attrValueContent   = "content"
	attrValueEmail     = "email"
	attrValueEven      = "even"
	attrValueFont      = "font"
	attrValueFooter    = "footer"
	attrValueHeader    = "header"
	attrValueHover     = "hover"
	attrValueImage     = "image"
	attrValueNoopener  = "noopener"
	attrValueNoHeader  = "noheader"
	attrValueNone      = "none"
	attrValueOdd       = "odd"
	attrValueRevDate   = "revdate"
	attrValueRevNumber = "revnumber"
	attrValueRows      = "rows"
	attrValueSides     = "sides"
	attrValueTitle     = attrNameTitle
	attrValueTopbot    = "topbot"
)

const (
	classNameArabic       = "arabic"
	classNameChecklist    = "checklist"
	classNameFitContent   = "fit-content"
	classNameFrameAll     = "frame-all"
	classNameFrameEnds    = "frame-ends"
	classNameFrameNone    = "frame-none"
	classNameFrameSides   = "frame-sides"
	classNameGridAll      = "grid-all"
	classNameGridCols     = "grid-cols"
	classNameGridNone     = "grid-none"
	classNameGridRows     = "grid-rows"
	classNameLoweralpha   = "loweralpha"
	classNameLowerroman   = "lowerroman"
	classNameStretch      = "stretch"
	classNameStripesAll   = "stripes-all"
	classNameStripesEven  = "stripes-even"
	classNameStripesHover = "stripes-hover"
	classNameStripesOdd   = "stripes-odd"
	classNameTableblock   = "tableblock"
	classNameUpperalpha   = "upperalpha"
	classNameUpperroman   = "upperroman"
)

// List of document metadata.
const (
	metaNameAuthor         = attrValueAuthor
	metaNameAuthorInitials = "authorinitials"
	metaNameDescription    = "description"
	metaNameDocTitle       = "doctitle"
	metaNameEmail          = attrValueEmail
	metaNameFirstName      = "firstname"
	metaNameIDPrefix       = "idprefix"
	metaNameIDSeparator    = "idseparator"
	metaNameKeywords       = "keywords"
	metaNameLastName       = "lastname"
	metaNameMiddleName     = "middlename"
	metaNameNoFooter       = "nofooter"
	metaNameNoHeader       = "noheader"
	metaNameNoHeaderFooter = "no-header-footer"
	metaNameNoTitle        = "notitle"
	metaNameRevDate        = "revdate"
	metaNameRevNumber      = "revnumber"
	metaNameRevRemark      = "revremark"
	metaNameSectAnchors    = "sectanchors"
	metaNameSectIDs        = "sectids"
	metaNameSectLinks      = "sectlinks"
	metaNameSectNumLevel   = "sectnumlevels"
	metaNameSectNums       = "sectnums"
	metaNameShowTitle      = "showtitle"
	metaNameTOC            = "toc"
	metaNameTOCLevels      = "toclevels"
	metaNameTOCTitle       = "toc-title"
	metaNameTableCaption   = "table-caption"
	metaNameTitle          = attrNameTitle
	metaNameTitleSeparator = "title-separator"
	metaNameVersionLabel   = "version-label"
)

// List of possible metadata value.
const (
	metaValueAuto     = "auto"
	metaValueMacro    = "macro"
	metaValuePreamble = "preamble"
	metaValueLeft     = "left"
	metaValueRight    = "right"
)

const (
	optNameAutoplay               = "autoplay"
	optNameAutowidth              = "autowidth"
	optNameControls               = "controls"
	optNameLoop                   = "loop"
	optNameNocontrols             = "nocontrols"
	optVideoFullscreen            = "fs"
	optVideoModest                = "modest"
	optVideoNofullscreen          = "nofullscreen"
	optVideoPlaylist              = "playlist"
	optVideoYoutubeModestbranding = "modestbranding"
)

const (
	prefixInclude = "include::"
)

const (
	admonitionCaution   = "CAUTION"
	admonitionImportant = "IMPORTANT"
	admonitionNote      = "NOTE"
	admonitionTip       = "TIP"
	admonitionWarning   = "WARNING"
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
	symbolQuoteDoubleBegin = "&#8220;"
	symbolQuoteDoubleEnd   = "&#8221;"
	symbolQuoteSingleBegin = "&#8216;"
	symbolQuoteSingleEnd   = "&#8217;"
	symbolChecked          = "&#10003;"
	symbolUnchecked        = "&#10063;"
)

var adocStyles map[string]int64 = map[string]int64{
	"colophon":          styleSectionColophon,
	"abstract":          styleSectionAbstract,
	"preface":           styleSectionPreface,
	"dedication":        styleSectionDedication,
	attrNameDiscrete:    styleSectionDiscrete,
	"partintro":         styleSectionPartIntroduction,
	"appendix":          styleSectionAppendix,
	"glossary":          styleSectionGlossary,
	"bibliography":      styleSectionBibliography,
	"index":             styleSectionIndex,
	".lead":             styleParagraphLead,
	".normal":           styleParagraphNormal,
	"arabic":            styleNumberingArabic,
	"decimal":           styleNumberingDecimal,
	"loweralpha":        styleNumberingLoweralpha,
	"upperalpha":        styleNumberingUpperalpha,
	"lowerroman":        styleNumberingLowerroman,
	"upperroman":        styleNumberingUpperroman,
	"lowergreek":        styleNumberingLowergreek,
	"horizontal":        styleDescriptionHorizontal,
	"qanda":             styleDescriptionQandA,
	admonitionCaution:   styleAdmonition,
	admonitionImportant: styleAdmonition,
	admonitionNote:      styleAdmonition,
	admonitionTip:       styleAdmonition,
	admonitionWarning:   styleAdmonition,
	"listing":           styleBlockListing,
	"quote":             styleQuote,
	"source":            styleSource,
	"verse":             styleVerse,
}

var _attrRef map[string]string = map[string]string{
	"amp":            "&",
	"apos":           htmlSymbolSingleQuote, // '
	"asterisk":       "*",
	"backslash":      `\`,
	"backtick":       "`",
	"blank":          "",
	"brvbar":         htmlSymbolBrokenVerticalBar, // ¦
	"caret":          "^",
	"cpp":            "C++",
	"deg":            htmlSymbolDegreeSign, // °
	"empty":          "",
	"endsb":          "]",
	"gt":             ">",
	"ldquo":          htmlSymbolLeftDoubleQuote,
	"lsquo":          htmlSymbolLeftSingleQuote,
	"lt":             "<",
	"nbsp":           htmlSymbolNonBreakingSpace,
	"plus":           htmlSymbolPlus,
	"quot":           htmlSymbolDoubleQuote,
	"rdquo":          htmlSymbolRightDoubleQuote,
	"rsquo":          htmlSymbolRightSingleQuote,
	"sp":             " ",
	"startsb":        "[",
	"tilde":          "~",
	"two-colons":     "::",
	"two-semicolons": ";;",
	"vbar":           "|",
	"wj":             htmlSymbolWordJoiner,
	"zwsp":           htmlSymbolZeroWidthSpace,
}

func applySubstitutions(doc *Document, content []byte) []byte {
	var (
		raw    = bytes.TrimRight(content, " \n")
		newraw = make([]byte, 0, len(raw))
		buf    = bytes.NewBuffer(newraw)
		c      byte
		x      int
	)
	for x < len(raw) {
		c = raw[x]
		if c == '{' {
			newRaw, ok := parseAttrRef(doc, raw, x)
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

func generateID(doc *Document, str string) string {
	idPrefix := "_"
	v, ok := doc.Attributes[metaNameIDPrefix]
	if ok {
		idPrefix = strings.TrimSpace(v)
	}

	idSep := "_"
	v, ok = doc.Attributes[metaNameIDSeparator]
	if ok {
		idSep = strings.TrimSpace(v)
	}

	id := make([]rune, 0, len(str)+1)
	id = append(id, []rune(idPrefix)...)
	for _, c := range strings.ToLower(str) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			id = append(id, c)
		} else {
			if id[len(id)-1] != '_' {
				id = append(id, []rune(idSep)...)
			}
		}
	}
	return strings.TrimRight(string(id), idSep)
}

func isAdmonition(line []byte) bool {
	var x int
	if bytes.HasPrefix(line, []byte(admonitionCaution)) {
		x = len(admonitionCaution)
	} else if bytes.HasPrefix(line, []byte(admonitionImportant)) {
		x = len(admonitionImportant)
	} else if bytes.HasPrefix(line, []byte(admonitionNote)) {
		x = len(admonitionNote)
	} else if bytes.HasPrefix(line, []byte(admonitionTip)) {
		x = len(admonitionTip)
	} else if bytes.HasPrefix(line, []byte(admonitionWarning)) {
		x = len(admonitionWarning)
	} else {
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
	_, x := indexUnescape(line, []byte(":: "))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(line, []byte("::\t"))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(line, []byte("::"))
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

//
// isValidID will return true if id is valid XML ID, where the first character
// is '-', '_', or letter; and the rest is either '-', '_', letter or digits.
//
func isValidID(id []byte) bool {
	for x, r := range string(id) {
		if x == 0 {
			if !(r == ':' || r == '-' || r == '_' || unicode.IsLetter(r)) {
				return false
			}
			continue
		}
		if r == ':' || r == '-' || r == '_' || r == '.' ||
			unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

//
// parseAttribute parse document attribute and return its key and optional
// value.
//
//	DOC_ATTRIBUTE  = ":" DOC_ATTR_KEY ":" *STRING LF
//
//	DOC_ATTR_KEY   = ( "toc" / "sectanchors" / "sectlinks"
//	               /   "imagesdir" / "data-uri" / *META_KEY ) LF
//
//	META_KEY_CHAR  = (A..Z | a..z | 0..9 | '_')
//
//	META_KEY       = 1META_KEY_CHAR *(META_KEY_CHAR | '-')
//
func parseAttribute(line []byte, strict bool) (key, value string, ok bool) {
	var sb strings.Builder

	if !(ascii.IsAlnum(line[1]) || line[1] == '_') {
		return "", "", false
	}

	sb.WriteByte(line[1])
	x := 2
	for ; x < len(line); x++ {
		if line[x] == ':' {
			break
		}
		if ascii.IsAlnum(line[x]) || line[x] == '_' ||
			line[x] == '-' || line[x] == '!' {
			sb.WriteByte(line[x])
			continue
		}
		if strict {
			return "", "", false
		}
	}
	if x == len(line) {
		return "", "", false
	}

	valb := bytes.TrimSpace(line[x+1:])

	return sb.String(), string(valb), true
}

//
// parseAttrRef parse the attribute reference, an attribute key wrapped by
// "{" "}".  If the attribute reference exist, replace the content with the
// attribute value and reset the parser state to zero.
//
func parseAttrRef(doc *Document, content []byte, x int) (
	newContent []byte, ok bool,
) {
	raw := content[x+1:]
	attrName, idx := indexByteUnescape(raw, '}')
	if idx < 0 {
		return nil, false
	}

	name := string(bytes.TrimSpace(bytes.ToLower(attrName)))
	attrValue, ok := _attrRef[name]
	if !ok {
		attrValue, ok = doc.Attributes[name]
		if !ok {
			return nil, false
		}

		// Add prefix "mailto:" if the ref name start with email, so
		// it can be parsed by caller as macro link.
		if name == "email" || strings.HasPrefix(name, "email_") {
			attrValue = "mailto:" + attrValue + "[" + attrValue + "]"
		}
	}

	rest := content[x+idx+2:]
	newContent = make([]byte, 0, len(attrValue)+len(rest))
	newContent = append(newContent, attrValue...)
	newContent = append(newContent, rest...)
	return newContent, true
}

//
// parseIDLabel parse the string "ID (,LABEL)" into ID and label.
// It will return empty id and label if ID is not valid.
//
func parseIDLabel(s []byte) (id, label []byte) {
	idLabel := bytes.Split(s, []byte(","))
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
	pi := newInlineParser(doc, content)
	pi.do()
	return pi.container
}

//
// parseStyle get the style based on string value.
//
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

//
// whatKindOfLine return the kind of line.
// It will return lineKindText if the line does not match with known syntax.
//
func whatKindOfLine(line []byte) (kind int, spaces, got []byte) {
	kind = lineKindText
	if len(line) == 0 {
		return lineKindEmpty, nil, line
	}
	if bytes.HasPrefix(line, []byte("////")) {
		// Check for comment block first, since we use HasPrefix to
		// check for single line comment.
		return lineKindBlockComment, spaces, line
	}
	if bytes.HasPrefix(line, []byte("//")) {
		// Use HasPrefix to allow single line comment without space,
		// for example "//comment".
		return lineKindComment, spaces, line
	}
	if bytes.Equal(line, []byte("'''")) ||
		bytes.Equal(line, []byte("---")) ||
		bytes.Equal(line, []byte("- - -")) ||
		bytes.Equal(line, []byte("***")) ||
		bytes.Equal(line, []byte("* * *")) {
		return lineKindHorizontalRule, spaces, line
	}
	if bytes.Equal(line, []byte("<<<")) {
		return lineKindPageBreak, spaces, line
	}
	if bytes.Equal(line, []byte("--")) {
		return elKindBlockOpen, spaces, line
	}
	if bytes.Equal(line, []byte("____")) {
		return elKindBlockExcerpts, spaces, line
	}
	if bytes.Equal(line, []byte("....")) {
		return elKindBlockLiteral, nil, line
	}
	if bytes.Equal(line, []byte("++++")) {
		return elKindBlockPassthrough, spaces, line
	}
	if bytes.Equal(line, []byte("****")) {
		return elKindBlockSidebar, nil, line
	}
	if bytes.Equal(line, []byte("====")) {
		return elKindBlockExample, spaces, line
	}

	if bytes.HasPrefix(line, []byte("|===")) {
		return elKindTable, nil, line
	}

	if bytes.Equal(line, []byte("[listing]")) {
		return elKindBlockListingNamed, nil, line
	}
	if bytes.Equal(line, []byte("[literal]")) {
		return elKindBlockLiteralNamed, nil, line
	}
	if bytes.Equal(line, []byte("toc::[]")) {
		return elKindMacroTOC, spaces, line
	}
	if bytes.HasPrefix(line, []byte("image::")) {
		return elKindBlockImage, spaces, line
	}
	if bytes.HasPrefix(line, []byte("include::")) {
		return lineKindInclude, nil, line
	}
	if bytes.HasPrefix(line, []byte("video::")) {
		return elKindBlockVideo, nil, line
	}
	if bytes.HasPrefix(line, []byte("audio::")) {
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

		// A line idented with space only allowed on list item,
		// otherwise it would be set as literal paragraph.

		if isLineDescriptionItem(line) {
			return elKindListDescriptionItem, spaces, line
		}

		if line[0] != '*' && line[0] != '.' {
			return elKindLiteralParagraph, spaces, line
		}
	}

	if line[0] == ':' {
		kind = lineKindAttribute
	} else if line[0] == '[' {
		newline := bytes.TrimRight(line, " \t")
		l := len(newline)
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
	} else if line[0] == '=' {
		subs := bytes.Fields(line)
		if bytes.Equal(subs[0], []byte("==")) {
			kind = elKindSectionL1
		} else if bytes.Equal(subs[0], []byte("===")) {
			kind = elKindSectionL2
		} else if bytes.Equal(subs[0], []byte("====")) {
			kind = elKindSectionL3
		} else if bytes.Equal(subs[0], []byte("=====")) {
			kind = elKindSectionL4
		} else if bytes.Equal(subs[0], []byte("======")) {
			kind = elKindSectionL5
		}
	} else if line[0] == '.' {
		if len(line) <= 1 {
			kind = lineKindText
		} else if ascii.IsAlnum(line[1]) {
			kind = lineKindBlockTitle
		} else {
			x := 0
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
	} else if line[0] == '*' {
		if len(line) <= 1 {
			kind = lineKindText
		} else {
			x := 0
			for ; x < len(line); x++ {
				if line[x] == '*' {
					continue
				}
				if line[x] == ' ' || line[x] == '\t' {
					kind = elKindListUnorderedItem
					return kind, spaces, line
				}
				kind = lineKindText
				return kind, spaces, line
			}
		}
	} else if bytes.Equal(line, []byte("+")) {
		kind = lineKindListContinue
	} else if bytes.Equal(line, []byte("----")) {
		kind = elKindBlockListing
	} else if isLineDescriptionItem(line) {
		kind = elKindListDescriptionItem
	}
	return kind, spaces, line
}
