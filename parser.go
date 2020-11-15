// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/parser"
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
	nodeKindParagraph                  // Wrapper.
	nodeKindLiteralParagraph           // 10: Line start with space
	nodeKindBlockAudio                 // "audio::"
	nodeKindBlockExample               // "===="
	nodeKindBlockExcerpts              // "____"
	nodeKindBlockImage                 // "image::"
	nodeKindBlockListing               // "----"
	nodeKindBlockListingNamed          // "[listing]"
	nodeKindBlockLiteral               // "...."
	nodeKindBlockLiteralNamed          // "[literal]"
	nodeKindBlockOpen                  // Block wrapped with "--"
	nodeKindBlockPassthrough           // 20: Block wrapped with "++++"
	nodeKindBlockSidebar               // "****"
	nodeKindBlockVideo                 // "video::"
	nodeKindCrossReference             // "<<" REF ("," LABEL) ">>"
	nodeKindInlineID                   // "[[" REF_ID "]]" TEXT
	nodeKindInlineIDShort              // "[#" REF_ID "]#" TEXT "#"
	nodeKindInlineImage                // Inline macro for "image:"
	nodeKindInlineParagraph            //
	nodeKindListOrdered                // Wrapper.
	nodeKindListOrderedItem            // Line start with ". "
	nodeKindListUnordered              // 30: Wrapper.
	nodeKindListUnorderedItem          // Line start with "* "
	nodeKindListDescription            // Wrapper.
	nodeKindListDescriptionItem        // Line that has "::" + WSP
	nodeKindMacroTOC                   // "toc::[]"
	nodeKindPassthrough                // Text wrapped inside "+"
	nodeKindPassthroughDouble          // Text wrapped inside "++"
	nodeKindPassthroughTriple          // Text wrapped inside "+++"
	nodeKindSymbolQuoteDoubleBegin     // The ("`)
	nodeKindSymbolQuoteDoubleEnd       // The (`")
	nodeKindSymbolQuoteSingleBegin     // 40: The ('`)
	nodeKindSymbolQuoteSingleEnd       // The (`')
	nodeKindText                       //
	nodeKindTextBold                   // Text wrapped by "*"
	nodeKindTextItalic                 // Text wrapped by "_"
	nodeKindTextMono                   // Text wrapped by "`"
	nodeKindTextSubscript              // Word wrapped by '~'
	nodeKindTextSuperscript            // Word wrapped by '^'
	nodeKindUnconstrainedBold          // Text wrapped by "**"
	nodeKindUnconstrainedItalic        // Text wrapped by "__"
	nodeKindUnconstrainedMono          // 50: Text wrapped by "``"
	nodeKindURL                        // Anchor text.
	lineKindAdmonition                 // "LABEL: WSP"
	lineKindAttribute                  // ":" ATTR_NAME ":" (ATTR_VALUE)
	lineKindAttributeElement           // "[" ATTR_NAME ("=" ATTR_VALUE)"]"
	lineKindBlockComment               // Block start and end with "////"
	lineKindBlockTitle                 // Line start with ".<alnum>"
	lineKindComment                    // Line start with "//"
	lineKindEmpty                      // LF
	lineKindHorizontalRule             // "'''", "---", "- - -", "***", "* * *"
	lineKindID                         // 60: "[[" REF_ID "]]"
	lineKindIDShort                    // "[#" REF_ID "]#" TEXT "#"
	lineKindListContinue               // "+" LF
	lineKindPageBreak                  // "<<<"
	lineKindStyleClass                 // "[.x.y]"
	lineKindText                       // 1*VCHAR
)

const (
	attrNameAlign       = "align"
	attrNameAlt         = "alt"
	attrNameEnd         = "end"
	attrNameFloat       = "float"
	attrNameHeight      = "height"
	attrNameHref        = "href"
	attrNameIcons       = "icons"
	attrNameLang        = "lang"
	attrNameLink        = "link"
	attrNameOptions     = "options"
	attrNamePoster      = "poster"
	attrNameRefText     = "reftext"
	attrNameRel         = "rel"
	attrNameRole        = "role"
	attrNameSrc         = "src"
	attrNameStart       = "start"
	attrNameTarget      = "target"
	attrNameTheme       = "theme"
	attrNameTitle       = "title"
	attrNameVimeo       = "vimeo"
	attrNameWidth       = "width"
	attrNameWindow      = "window"
	attrNameYoutube     = "youtube"
	attrNameYoutubeLang = "hl"
)

const (
	attrValueAttribution = "attribution"
	attrValueAuthor      = "author"
	attrValueBare        = "bare"
	attrValueBlank       = "_blank"
	attrValueContent     = "content"
	attrValueEmail       = "email"
	attrValueFont        = "font"
	attrValueImage       = "image"
	attrValueNoopener    = "noopener"
	attrValueRevDate     = "revdate"
	attrValueRevNumber   = "revnumber"
	attrValueTitle       = attrNameTitle
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
	styleSectionPreface
	styleSectionDedication
	styleSectionPartIntroduction
	styleSectionAppendix
	styleSectionGlossary
	styleSectionBibliography
	styleSectionIndex
	styleParagraphLead
	styleParagraphNormal
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
)

var adocStyles map[string]int64 = map[string]int64{
	"colophon":          styleSectionColophon,
	"abstract":          styleSectionAbstract,
	"preface":           styleSectionPreface,
	"dedication":        styleSectionDedication,
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

func isAdmonition(line string) bool {
	var x int
	if strings.HasPrefix(line, admonitionCaution) {
		x = len(admonitionCaution)
	} else if strings.HasPrefix(line, admonitionImportant) {
		x = len(admonitionImportant)
	} else if strings.HasPrefix(line, admonitionNote) {
		x = len(admonitionNote)
	} else if strings.HasPrefix(line, admonitionTip) {
		x = len(admonitionTip)
	} else if strings.HasPrefix(line, admonitionWarning) {
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

func isLineDescriptionItem(line string) bool {
	bline := []byte(line)
	_, x := indexUnescape(bline, []byte(":: "))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(bline, []byte("::\t"))
	if x > 0 {
		return true
	}
	_, x = indexUnescape(bline, []byte("::"))
	return x > 0
}

// isRefTitle will return true if one of character is upper case or white
// space.
func isRefTitle(s string) bool {
	for _, r := range s {
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

func isTitle(line string) bool {
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
func isValidID(id string) bool {
	for x, r := range id {
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
func parseAttribute(line string, strict bool) (key, value string, ok bool) {
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

	return sb.String(), strings.TrimSpace(line[x+1:]), true
}

//
// parseAttributeElement parse list of attributes in between "[" "]".
//
//	BLOCK_ATTRS = BLOCK_ATTR *("," BLOCK_ATTR)
//
//	BLOCK_ATTR  = ATTR_NAME ( "=" (DQUOTE) ATTR_VALUE (DQUOTE) )
//
//	ATTR_NAME   = WORD
//
//	ATTR_VALUE  = STRING
//
// The attribute may not have a value.
//
// If the attribute value contains space or comma, it must be wrapped with
// double quote.
// The double quote on value will be removed when stored on opts.
//
// It will return nil if input is not a valid block attribute.
//
func parseAttributeElement(in string) (attrName, attrValue string, opts []string) {
	p := parser.New(in, `[,="]`)
	tok, c := p.Token()
	if c != '[' {
		return "", "", nil
	}
	if len(tok) > 0 {
		return "", "", nil
	}

	for c != 0 {
		tok, c = p.Token()
		tok = strings.TrimSpace(tok)
		if c == '"' && len(tok) == 0 {
			tok, c = p.ReadEnclosed('"', '"')
			tok = strings.TrimSpace(tok)
			opts = append(opts, tok)
			continue
		}
		if c == ',' || c == ']' {
			if len(tok) > 0 {
				opts = append(opts, tok)
			}
			if c == ']' {
				break
			}
			continue
		}
		if c != '=' {
			// Ignore invalid attribute.
			for c != ',' && c != 0 {
				_, c = p.Token()
			}
			continue
		}
		key := tok
		tok, c = p.Token()
		tok = strings.TrimSpace(tok)
		if c == '"' {
			tok, c = p.ReadEnclosed('"', '"')
			tok = strings.TrimSpace(tok)
			opts = append(opts, key+"="+tok)
		} else {
			opts = append(opts, key+"="+tok)
		}

		for c != ',' && c != 0 {
			_, c = p.Token()
		}
	}
	if len(opts) == 0 {
		return "", "", nil
	}

	nameValue := strings.Split(opts[0], "=")
	attrName = nameValue[0]
	if len(nameValue) >= 2 {
		attrValue = strings.Join(nameValue[1:], "=")
	}

	return attrName, attrValue, opts
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

	attrName = bytes.TrimSpace(bytes.ToLower(attrName))

	attrValue, ok := doc.Attributes[string(attrName)]
	if !ok {
		return nil, false
	}

	rest := content[x+idx+2:]
	newContent = make([]byte, 0, len(attrValue)+len(rest))
	newContent = append(newContent, attrValue...)
	newContent = append(newContent, rest...)
	return newContent, true
}

//
// parseIDLabel parse the s "ID (,LABEL)" into ID and label.
// It will return empty id and label if ID is not valid.
//
func parseIDLabel(s string) (id, label string) {
	idLabel := strings.Split(s, ",")
	id = idLabel[0]
	if len(idLabel) >= 2 {
		label = idLabel[1]
	}
	if isValidID(idLabel[0]) {
		return id, label
	}
	return "", ""
}

func parseInlineMarkup(doc *Document, content []byte) (container *adocNode) {
	pi := newParserInline(doc, content)
	pi.do()
	return pi.container
}

//
// parseStyle parse line that start with "[" and end with "]".
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
func whatKindOfLine(line string) (kind int, spaces, got string) {
	kind = lineKindText
	if len(line) == 0 {
		return lineKindEmpty, spaces, line
	}
	if strings.HasPrefix(line, "////") {
		// Check for comment block first, since we use HasPrefix to
		// check for single line comment.
		return lineKindBlockComment, spaces, line
	}
	if strings.HasPrefix(line, "//") {
		// Use HasPrefix to allow single line comment without space,
		// for example "//comment".
		return lineKindComment, spaces, line
	}
	if line == "'''" || line == "---" || line == "- - -" ||
		line == "***" || line == "* * *" {
		return lineKindHorizontalRule, spaces, line
	}
	if line == "<<<" {
		return lineKindPageBreak, spaces, line
	}
	if line == "--" {
		return nodeKindBlockOpen, spaces, line
	}
	if line == "____" {
		return nodeKindBlockExcerpts, spaces, line
	}
	if line == "...." {
		return nodeKindBlockLiteral, "", line
	}
	if line == "++++" {
		return nodeKindBlockPassthrough, spaces, line
	}
	if line == "[listing]" {
		return nodeKindBlockListingNamed, "", line
	}
	if line == "[literal]" {
		return nodeKindBlockLiteralNamed, "", line
	}
	if line == "toc::[]" {
		return nodeKindMacroTOC, spaces, line
	}
	if strings.HasPrefix(line, "image::") {
		return nodeKindBlockImage, spaces, line
	}
	if strings.HasPrefix(line, "video::") {
		line = strings.TrimRight(line[7:], " \t")
		return nodeKindBlockVideo, "", line
	}
	if strings.HasPrefix(line, "audio::") {
		line = strings.TrimRight(line[7:], " \t")
		return nodeKindBlockAudio, "", line
	}
	if isAdmonition(line) {
		return lineKindAdmonition, "", line
	}

	var (
		x        int
		r        rune
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
		newline := strings.TrimRight(line, " \t")
		l := len(newline)
		if newline[l-1] != ']' {
			return lineKindText, "", line
		}
		if l >= 5 {
			if newline[1] == '[' && newline[l-2] == ']' {
				return lineKindID, "", line
			}
		}
		if l >= 4 {
			if line[1] == '#' {
				return lineKindIDShort, "", line
			}
			if line[1] == '.' {
				return lineKindStyleClass, "", line
			}
		}
		return lineKindAttributeElement, spaces, line
	} else if line[0] == '=' {
		if line == "====" {
			return nodeKindBlockExample, spaces, line
		}

		subs := strings.Fields(line)
		switch subs[0] {
		case "==":
			kind = nodeKindSectionL1
		case "===":
			kind = nodeKindSectionL2
		case "====":
			kind = nodeKindSectionL3
		case "=====":
			kind = nodeKindSectionL4
		case "======":
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
		} else if line == "****" {
			kind = nodeKindBlockSidebar
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
	} else if line == "+" {
		kind = lineKindListContinue
	} else if line == "----" {
		kind = nodeKindBlockListing
	} else if isLineDescriptionItem(line) {
		kind = nodeKindListDescriptionItem
	}
	return kind, spaces, line
}
