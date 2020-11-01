// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
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
	lineKindAttribute                  // Line start with ":"
	lineKindBlockComment               // Block start and end with "////"
	lineKindBlockTitle                 // Line start with ".<alnum>"
	lineKindComment                    // Line start with "//"
	lineKindEmpty                      // LF
	lineKindHorizontalRule             // "'''", "---", "- - -", "***", "* * *"
	lineKindID                         // "[[" REF_ID "]]"
	lineKindIDShort                    // 60:"[#" REF_ID "]#" TEXT "#"
	lineKindListContinue               // A single "+" line
	lineKindPageBreak                  // "<<<"
	lineKindStyle                      // Line start with "["
	lineKindStyleClass                 // Custom style "[.x.y]"
	lineKindText                       //
)

const (
	attrNameAlign       = "align"
	attrNameAlt         = "alt"
	attrNameEnd         = "end"
	attrNameFloat       = "float"
	attrNameHeight      = "height"
	attrNameHref        = "href"
	attrNameID          = "id"
	attrNameLang        = "lang"
	attrNameOptions     = "options"
	attrNamePoster      = "poster"
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
	attrValueBlank    = "_blank"
	attrValueNoopener = "noopener"
)

// List of document metadata.
const (
	metaNameTOC       = "toc"
	metaNameTOCLevels = "toclevels"
	metaNameTOCTitle  = "toc-title"
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
	styleNone            int64 = iota
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
	styleVerse
	styleTextBold
	styleTextItalic
	styleTextMono
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

func generateID(str string) string {
	id := make([]rune, 0, len(str)+1)
	id = append(id, '_')
	for _, c := range strings.ToLower(str) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			id = append(id, c)
		} else {
			if id[len(id)-1] != '_' {
				id = append(id, '_')
			}
		}
	}
	return strings.TrimRight(string(id), "_")
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
	x := strings.Index(line, ":: ")
	if x > 0 {
		return true
	}
	x = strings.Index(line, "::\t")
	if x > 0 {
		return true
	}
	x = strings.Index(line, "::")
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
		if line[1] == ' ' || line[1] == '\t' {
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
// parseBlockAttribute parse list of attributes in between "[" "]".
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
// The double quote on value will be removed when stored on output.
//
// It will return nil if input is not a valid block attribute.
//
func parseBlockAttribute(in string) (out []string) {
	p := parser.New(in, `[,="]`)
	tok, c := p.Token()
	if c != '[' {
		return nil
	}
	if len(tok) > 0 {
		return nil
	}

	for c != 0 {
		tok, c = p.Token()
		tok = strings.TrimSpace(tok)
		if c == '"' && len(tok) == 0 {
			tok, c = p.ReadEnclosed('"', '"')
			tok = strings.TrimSpace(tok)
			out = append(out, tok)
			continue
		}
		if c == ',' || c == ']' {
			if len(tok) > 0 {
				out = append(out, tok)
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
			out = append(out, key+"="+tok)
		} else {
			out = append(out, key+"="+tok)
		}

		for c != ',' && c != 0 {
			_, c = p.Token()
		}
	}
	return out
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
func parseStyle(line string) (styleName string, styleKind int64, opts []string) {
	line = strings.Trim(line, "[]")
	parts := strings.Split(line, ",")
	styleName = strings.Trim(parts[0], "\"")

	// Check for admonition label first...
	styleKind = adocStyles[styleName]
	if styleKind > 0 {
		return styleName, styleKind, parts[1:]
	}

	styleName = strings.ToLower(styleName)
	styleKind = adocStyles[styleName]

	return styleName, styleKind, parts[1:]
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
		return lineKindStyle, spaces, line
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
