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
	nodeKindUnknown                int = iota
	nodeKindDocHeader                  // Wrapper.
	nodeKindPreamble                   // Wrapper.
	nodeKindDocContent                 // Wrapper.
	nodeKindSectionL1                  // Line started with "=="
	nodeKindSectionL2                  // 5: Line started with "==="
	nodeKindSectionL3                  // Line started with "===="
	nodeKindSectionL4                  // Line started with "====="
	nodeKindSectionL5                  // Line started with "======"
	nodeKindSectionDiscrete            // "[discrete]"
	nodeKindParagraph                  // 10: Wrapper.
	nodeKindLiteralParagraph           // Line start with space
	nodeKindBlockAudio                 // "audio::"
	nodeKindBlockExample               // "===="
	nodeKindBlockExcerpts              // "____"
	nodeKindBlockImage                 // "image::"
	nodeKindBlockListing               // "----"
	nodeKindBlockListingNamed          // "[listing]"
	nodeKindBlockLiteral               // "...."
	nodeKindBlockLiteralNamed          // "[literal]"
	nodeKindBlockOpen                  // 20: Block wrapped with "--"
	nodeKindBlockPassthrough           // Block wrapped with "++++"
	nodeKindBlockSidebar               // "****"
	nodeKindBlockVideo                 // "video::"
	nodeKindCrossReference             // "<<" REF ("," LABEL) ">>"
	nodeKindInlineID                   // "[[" REF_ID "]]" TEXT
	nodeKindInlineIDShort              // "[#" REF_ID "]#" TEXT "#"
	nodeKindInlineImage                // Inline macro for "image:"
	nodeKindInlineParagraph            //
	nodeKindListOrdered                // Wrapper.
	nodeKindListOrderedItem            // 30: Line start with ". "
	nodeKindListUnordered              // Wrapper.
	nodeKindListUnorderedItem          // Line start with "* "
	nodeKindListDescription            // Wrapper.
	nodeKindListDescriptionItem        // Line that has "::" + WSP
	nodeKindMacroTOC                   // "toc::[]"
	nodeKindPassthrough                // Text wrapped inside "+"
	nodeKindPassthroughDouble          // Text wrapped inside "++"
	nodeKindPassthroughTriple          // Text wrapped inside "+++"
	nodeKindSymbolQuoteDoubleBegin     // The ("`)
	nodeKindSymbolQuoteDoubleEnd       // 40: The (`")
	nodeKindSymbolQuoteSingleBegin     // The ('`)
	nodeKindSymbolQuoteSingleEnd       // The (`')
	nodeKindTable                      // "|==="
	nodeKindText                       //
	nodeKindTextBold                   // Text wrapped by "*"
	nodeKindTextItalic                 // Text wrapped by "_"
	nodeKindTextMono                   // Text wrapped by "`"
	nodeKindTextSubscript              // Word wrapped by '~'
	nodeKindTextSuperscript            // Word wrapped by '^'
	nodeKindUnconstrainedBold          // 50: Text wrapped by "**"
	nodeKindUnconstrainedItalic        // Text wrapped by "__"
	nodeKindUnconstrainedMono          // Text wrapped by "``"
	nodeKindURL                        // Anchor text.
	lineKindAdmonition                 // "LABEL: WSP"
	lineKindAttribute                  // ":" ATTR_NAME ":" (ATTR_VALUE)
	lineKindAttributeElement           // "[" ATTR_NAME ("=" ATTR_VALUE)"]"
	lineKindBlockComment               // Block start and end with "////"
	lineKindBlockTitle                 // Line start with ".<alnum>"
	lineKindComment                    // Line start with "//"
	lineKindEmpty                      // 60: LF
	lineKindHorizontalRule             // "'''", "---", "- - -", "***", "* * *"
	lineKindID                         // "[[" REF_ID "]]"
	lineKindIDShort                    // "[#" REF_ID "]#" TEXT "#"
	lineKindInclude                    // "include::"
	lineKindListContinue               // "+" LF
	lineKindPageBreak                  // "<<<"
	lineKindStyleClass                 // "[.x.y]"
	lineKindText                       // 1*VCHAR
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
	classNameChecklist    = "checklist"
	classNameFitContent   = "fit-content"
	classNameFrameAll     = "frame-all"
	classNameFrameEnds    = "frame-ends"
	classNameFrameSides   = "frame-sides"
	classNameFrameNone    = "frame-none"
	classNameGridAll      = "grid-all"
	classNameGridCols     = "grid-cols"
	classNameGridNone     = "grid-none"
	classNameGridRows     = "grid-rows"
	classNameStretch      = "stretch"
	classNameStripesAll   = "stripes-all"
	classNameStripesEven  = "stripes-even"
	classNameStripesHover = "stripes-hover"
	classNameStripesOdd   = "stripes-odd"
	classNameTableblock   = "tableblock"
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
	"verse":             styleVerse,
}

var _attrRef map[string]string = map[string]string{
	"vbar": "|",
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

// isRefTitle will return true if one of character is upper case or white
// space.
func isRefTitle(s []byte) bool {
	for _, r := range string(s) {
		if unicode.IsUpper(r) || unicode.IsSpace(r) {
			return true
		}
	}
	return false
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
			if !(r == '-' || r == '_' || unicode.IsLetter(r)) {
				return false
			}
			continue
		}
		if r == '-' || r == '_' || unicode.IsLetter(r) ||
			unicode.IsDigit(r) {
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

func parseInlineMarkup(doc *Document, content []byte) (container *adocNode) {
	pi := newParserInline(doc, content)
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
		return nodeKindBlockOpen, spaces, line
	}
	if bytes.Equal(line, []byte("____")) {
		return nodeKindBlockExcerpts, spaces, line
	}
	if bytes.Equal(line, []byte("....")) {
		return nodeKindBlockLiteral, nil, line
	}
	if bytes.Equal(line, []byte("++++")) {
		return nodeKindBlockPassthrough, spaces, line
	}
	if bytes.Equal(line, []byte("****")) {
		return nodeKindBlockSidebar, nil, line
	}
	if bytes.Equal(line, []byte("====")) {
		return nodeKindBlockExample, spaces, line
	}

	if bytes.HasPrefix(line, []byte("|===")) {
		return nodeKindTable, nil, line
	}

	if bytes.Equal(line, []byte("[listing]")) {
		return nodeKindBlockListingNamed, nil, line
	}
	if bytes.Equal(line, []byte("[literal]")) {
		return nodeKindBlockLiteralNamed, nil, line
	}
	if bytes.Equal(line, []byte("toc::[]")) {
		return nodeKindMacroTOC, spaces, line
	}
	if bytes.HasPrefix(line, []byte("image::")) {
		return nodeKindBlockImage, spaces, line
	}
	if bytes.HasPrefix(line, []byte("include::")) {
		return lineKindInclude, nil, line
	}
	if bytes.HasPrefix(line, []byte("video::")) {
		return nodeKindBlockVideo, nil, line
	}
	if bytes.HasPrefix(line, []byte("audio::")) {
		return nodeKindBlockAudio, nil, line
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
			return nodeKindListDescriptionItem, spaces, line
		}

		if line[0] != '*' && line[0] != '.' {
			return nodeKindLiteralParagraph, spaces, line
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
			kind = nodeKindSectionL1
		} else if bytes.Equal(subs[0], []byte("===")) {
			kind = nodeKindSectionL2
		} else if bytes.Equal(subs[0], []byte("====")) {
			kind = nodeKindSectionL3
		} else if bytes.Equal(subs[0], []byte("=====")) {
			kind = nodeKindSectionL4
		} else if bytes.Equal(subs[0], []byte("======")) {
			kind = nodeKindSectionL5
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
					kind = nodeKindListOrderedItem
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
					kind = nodeKindListUnorderedItem
					return kind, spaces, line
				}
				kind = lineKindText
				return kind, spaces, line
			}
		}
	} else if bytes.Equal(line, []byte("+")) {
		kind = lineKindListContinue
	} else if bytes.Equal(line, []byte("----")) {
		kind = nodeKindBlockListing
	} else if isLineDescriptionItem(line) {
		kind = nodeKindListDescriptionItem
	}
	return kind, spaces, line
}
