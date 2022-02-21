// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"strconv"

	"github.com/shuLhan/share/lib/ascii"
)

type cellFormat struct {
	ndupCol  int
	nspanCol int
	nspanRow int
	alignHor int
	alignVer int
	style    int
}

//
// parseCellFormat parse the cell format and return it.
// In case one of the character is not a valid format, it will return nil,
// considering the whole raw string as not cell format.
//
func parseCellFormat(raw string) (cf *cellFormat) {
	var (
		x     int
		n     int
		isDot bool
	)
	if len(raw) == 0 {
		return nil
	}

	cf = &cellFormat{}
	if ascii.IsDigit(raw[0]) {
		n, raw = parseCellFormatDigits(raw)
		if len(raw) == 0 {
			return nil
		}
		if raw[0] == '*' {
			cf.ndupCol = n
			x = 1
		} else if raw[0] == '+' {
			cf.nspanCol = n
			x = 1
		} else if raw[0] == '.' {
			cf.nspanCol = n
			n, raw = parseCellFormatDigits(raw[1:])
			if n == 0 {
				// Invalid format, there should be digits
				// after '.'
				return nil
			}
			if raw[0] != '+' {
				return nil
			}
			x = 1
			cf.nspanRow = n
		} else {
			return nil
		}
	}
	for ; x < len(raw); x++ {
		switch raw[x] {
		case '.':
			isDot = true
		case '<':
			if isDot {
				cf.alignVer = colAlignTop
				isDot = false
			} else {
				cf.alignHor = colAlignTop
			}
		case '^':
			if isDot {
				cf.alignVer = colAlignMiddle
				isDot = false
			} else {
				cf.alignHor = colAlignMiddle
			}
		case '>':
			if isDot {
				cf.alignVer = colAlignBottom
				isDot = false
			} else {
				cf.alignHor = colAlignBottom
			}
		case 'a', 'e', 'h', 'l', 'm', 's', 'v':
			cf.style = _colStyles[raw[x]]
		default:
			if isDot && ascii.IsDigit(raw[x]) {
				n, raw = parseCellFormatDigits(raw[x:])
				if len(raw) == 0 {
					return nil
				}
				if raw[0] != '+' {
					return nil
				}
				cf.nspanRow = n
				x = 0
				isDot = false
				continue
			}
			return nil
		}
	}
	return cf
}

func parseCellFormatDigits(s string) (n int, rest string) {
	var (
		x int
		b []byte
	)
	for ; x < len(s); x++ {
		if !ascii.IsDigit(s[x]) {
			break
		}
		b = append(b, s[x])
	}
	n, _ = strconv.Atoi(string(b))
	rest = s[x:]
	return n, rest
}
