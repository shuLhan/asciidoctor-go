// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"strings"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/math/big"
)

const (
	colAlignTop    int = iota // '<'
	colAlignMiddle            // '^'
	colAlignBottom            // '>'
)

const (
	colStyleDefault    int = iota << 1 // 'd'
	colStyleAsciidoc                   // 'a'
	colStyleEmphasis                   // 'e'
	colStyleHeader                     // 'h'
	colStyleLiteral                    // 'l'
	colStyleMonospaced                 // 'm'
	colStyleStrong                     // 's'
	colStyleVerse                      // 'v'
)

var _colStyles = map[byte]int{
	'a': colStyleAsciidoc,
	'e': colStyleEmphasis,
	'h': colStyleHeader,
	'l': colStyleLiteral,
	'm': colStyleMonospaced,
	's': colStyleStrong,
	'v': colStyleVerse,
}

type columnFormat struct {
	width *big.Rat

	classes []string

	alignHor int
	alignVer int
	style    int

	// isDefault will true if its contains '*'.
	isDefault bool

	// isAutowidth will true if its contains '~'.
	isAutowidth bool
}

func newColumnFormat() *columnFormat {
	return &columnFormat{
		width: big.NewRat(1),
	}
}

func (f *columnFormat) htmlClasses() string {
	return strings.Join(f.classes, " ")
}

func (f *columnFormat) merge(other *columnFormat) {
	if other.width != nil {
		f.width = big.NewRat(other.width)
	}
	if other.alignHor != 0 {
		f.alignHor = other.alignHor
	}
	if other.alignVer != 0 {
		f.alignVer = other.alignVer
	}
	if other.style != 0 {
		f.style = other.style
	}
}

// parseColumnFormat parse single "cols" format value, for example "3*.>" or
// ".^3l".
func parseColumnFormat(s string) (ncols int, format *columnFormat) {
	format = newColumnFormat()
	if len(s) == 0 {
		return 0, format
	}

	if ascii.IsDigit(s[0]) {
		format.width, s = parseColumnDigits(s)
	}
	var isAlignVertical bool
	for x := 0; x < len(s); x++ {
		switch s[x] {
		case '*':
			format.isDefault = true
			ncols = int(format.width.Int64())
			format.width = big.NewRat(1)
		case '.':
			isAlignVertical = true
		case '<':
			if isAlignVertical {
				format.alignVer = colAlignTop
				isAlignVertical = false
			} else {
				format.alignHor = colAlignTop
			}
		case '^':
			if isAlignVertical {
				format.alignVer = colAlignMiddle
				isAlignVertical = false
			} else {
				format.alignHor = colAlignMiddle
			}
		case '>':
			if isAlignVertical {
				format.alignVer = colAlignBottom
				isAlignVertical = false
			} else {
				format.alignHor = colAlignBottom
			}
		case '~':
			format.isAutowidth = true
		case 'a', 'e', 'h', 'l', 'm', 's', 'v':
			format.style = _colStyles[s[x]]
		default:
			if ascii.IsDigit(s[x]) {
				format.width, s = parseColumnDigits(s[x:])
			}
		}
	}
	return ncols, format
}

func parseColumnDigits(s string) (w *big.Rat, rest string) {
	var (
		x int
		n []byte
	)
	for ; x < len(s); x++ {
		if !ascii.IsDigit(s[x]) {
			break
		}
		n = append(n, s[x])
	}
	w = big.NewRat(string(n))
	rest = s[x:]
	return w, rest
}
