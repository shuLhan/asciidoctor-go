// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"strings"

	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
)

type elementAttribute struct {
	Attrs map[string]string

	ID       string
	rawStyle string

	roles   []string
	options []string

	style int64
	pos   int
}

func (ea *elementAttribute) addRole(role string) {
	role = strings.TrimSpace(role)
	if len(role) == 0 {
		return
	}
	ea.roles = libstrings.AppendUniq(ea.roles, role)
}

func (ea *elementAttribute) htmlClasses() string {
	if len(ea.roles) == 0 {
		return ``
	}
	return strings.Join(ea.roles, ` `)
}

// parseElementAttribute parse list of attributes in between `[` `]`.
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
// The double quote on value will be removed when stored on options.
func (ea *elementAttribute) parseElementAttribute(raw []byte) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return
	}
	if raw[0] != '[' {
		return
	}
	if raw[len(raw)-1] != ']' {
		return
	}
	raw = raw[1 : len(raw)-1]
	if len(raw) == 0 {
		return
	}
	var (
		c = raw[0]

		str       string
		buf       []byte
		prevc     byte
		x         int
		isEscaped bool
	)
	if c == '#' || c == '.' || c == '%' {
		prevc = c
		x = 1
	}
	for ; x < len(raw); x++ {
		c = raw[x]
		switch c {
		case '"':
			isEscaped = false
			x++
			for ; x < len(raw); x++ {
				if raw[x] == '\\' {
					if isEscaped {
						buf = append(buf, '\\')
						isEscaped = false
					} else {
						isEscaped = true
					}
					continue
				}
				if raw[x] == '"' {
					break
				}
				buf = append(buf, raw[x])
			}
		case ',':
			str = string(bytes.TrimSpace(buf))
			ea.setByPreviousChar(prevc, str)
			ea.pos++
			buf = buf[:0]
			prevc = c
		case '#', '%':
			ea.setByPreviousChar(prevc, string(bytes.TrimSpace(buf)))
			buf = buf[:0]
			prevc = c
		case '.':
			if ea.style == styleQuote || ea.style == styleVerse {
				// Make the '.' as part of attribution.
				if prevc == ',' {
					buf = append(buf, c)
					continue
				}
			} else if ea.style == styleLink && ea.pos == 0 {
				buf = append(buf, c)
				continue
			}
			ea.setByPreviousChar(prevc, string(bytes.TrimSpace(buf)))
			buf = buf[:0]
			prevc = c
		default:
			buf = append(buf, c)
		}
	}
	if len(buf) > 0 {
		ea.setByPreviousChar(prevc, string(bytes.TrimSpace(buf)))
	}
}

func (ea *elementAttribute) parseNamedValue(str string) {
	if ea.Attrs == nil {
		ea.Attrs = make(map[string]string)
	}

	var (
		kv  = strings.Split(str, `=`)
		key = kv[0]
		val = strings.TrimSpace(kv[1])
	)

	if len(val) == 0 {
		ea.Attrs[key] = ``
		return
	}
	if val[0] == '"' {
		val = val[1:]
	}
	if val[len(val)-1] == '"' {
		val = val[:len(val)-1]
	}

	var (
		rawvals = strings.Split(val, `,`)
		vals    = make([]string, 0, len(rawvals))

		v string
	)

	for _, v = range rawvals {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		vals = append(vals, v)
	}

	switch key {
	case attrNameOptions, attrNameOpts:
		ea.options = append(ea.options, vals...)
	case attrNameRole:
		ea.roles = append(ea.roles, vals...)
	default:
		ea.Attrs[key] = val
	}
}

func (ea *elementAttribute) setByPreviousChar(prevc byte, str string) {
	switch prevc {
	case 0:
		if strings.IndexByte(str, '=') > 0 {
			ea.parseNamedValue(str)
		} else {
			ea.rawStyle = str
			ea.style = parseStyle(str)
		}
	case '#':
		ea.ID = str
	case '.':
		ea.addRole(str)
	case '%':
		ea.options = append(ea.options, str)
	case ',':
		if strings.IndexByte(str, '=') > 0 {
			ea.parseNamedValue(str)
		} else {
			if ea.Attrs == nil {
				ea.Attrs = make(map[string]string)
			}

			switch ea.pos {
			case 1:
				switch ea.style {
				case styleQuote, styleVerse:
					ea.Attrs[attrNameAttribution] = str
				case styleSource:
					ea.Attrs[attrNameSource] = str
				default:
					ea.Attrs[str] = ``
				}
			case 2:
				switch ea.style {
				case styleQuote, styleVerse:
					ea.Attrs[attrNameCitation] = str
				default:
					ea.Attrs[str] = ``
				}
			}
		}
	}
}
