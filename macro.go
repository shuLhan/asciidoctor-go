// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"strconv"

	"github.com/shuLhan/share/lib/ascii"
)

const (
	macroFTP      = `ftp`
	macroFootnote = `footnote`
	macroHTTP     = `http`
	macroHTTPS    = `https`
	macroIRC      = `irc`
	macroImage    = `image`
	macroLink     = `link`
	macroMailto   = `mailto`
)

var (
	_macroKind = map[string]int{
		macroFTP:      elKindURL,
		macroFootnote: elKindFootnote,
		macroHTTP:     elKindURL,
		macroHTTPS:    elKindURL,
		macroIRC:      elKindURL,
		macroImage:    elKindInlineImage,
		macroLink:     elKindURL,
		macroMailto:   elKindURL,
	}
)

type macro struct {
	content *element

	// key represent the URL in elKindURL or elKindImage;
	// or ID in elKindFootnote.
	// For ID, it could be empty.
	key string

	// val represent the text for URL or image and footnote.
	rawContent []byte

	// level represent footnoted index number.
	level int
}

// parseMacroName parse inline macro.
//
// The parser read the content textBefore in backward order until it found one
// of macro name.
// Macro can be escaped using the backslash, for example "\link:", and it will
// be ignored.
//
// If macro name and value valid it will return the element for that macro.
func (pi *inlineParser) parseMacroName(textBefore []byte) (macroName string) {
	var (
		x int = len(textBefore) - 1

		ok bool
	)

	for x >= 0 {
		if !ascii.IsAlpha(textBefore[x]) {
			return ``
		}

		macroName = string(textBefore[x:])
		_, ok = _macroKind[macroName]
		if ok {
			break
		}

		x--
	}
	if !ok {
		return ``
	}

	x--
	if x >= 0 {
		if textBefore[x] == '\\' {
			// Macro is escaped.
			return ``
		}
	}

	return macroName
}

// parseMacroFootnote parse the footnote macro,
//
//	"footnote:" [ REF_ID ] "[" STRING "]"
//
// The text content does not have "footnote:".
//
// The REF_ID, reference the previous footnote defined with the same.
// The first footnote with REF_ID, should have the STRING defined.
// The next footnote with the same REF_ID, should not have the STRING
// defined; if its already defined, the STRING is ignored.
//
// It will return an element if footnote is valid.
func (pi *inlineParser) parseMacroFootnote(text []byte) (el *element, n int) {
	var (
		mcr    *macro
		id     string
		key    string
		vbytes []byte
		x      int
		exist  bool
	)

	vbytes, x = indexByteUnescape(text, '[')
	if x > 0 {
		if !isValidID(vbytes) {
			return nil, 0
		}
		id = string(vbytes)
	}

	n = x + 1
	text = text[n:]

	vbytes, x = indexByteUnescape(text, ']')
	if x < 0 {
		return nil, 0
	}

	n += x + 2

	mcr, exist = pi.doc.registerFootnote(id, vbytes)
	if exist {
		id = ``
		vbytes = nil
	} else {
		// Footnote without explicit ID will be set the key with its
		// level.
		key = strconv.FormatInt(int64(mcr.level), 10)
	}

	el = &element{
		key:   key,
		raw:   vbytes,
		kind:  elKindFootnote,
		level: mcr.level,

		elementAttribute: elementAttribute{
			ID: id,
		},
	}

	if vbytes != nil {
		mcr.content = parseInlineMarkup(pi.doc, vbytes)
	}

	return el, n
}
