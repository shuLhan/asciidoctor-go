// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// parseBlockImage parse the image block or line.
// The line parameter must not contains the "image::" block or "image:"
// macro prefix.
func (el *element) parseBlockImage(doc *Document, line []byte) bool {
	var attrBegin = bytes.IndexByte(line, '[')
	if attrBegin < 0 {
		return false
	}
	var attrEnd = bytes.IndexByte(line, ']')
	if attrEnd < 0 {
		return false
	}

	var link = bytes.TrimRight(line[:attrBegin], " \t")
	link = applySubstitutions(doc, link)

	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}
	el.Attrs[attrNameSrc] = string(link)

	var listAttribute = bytes.Split(line[attrBegin+1:attrEnd], []byte(`,`))
	if el.Attrs == nil {
		el.Attrs = make(map[string]string)
	}

	var (
		x        int
		battr    []byte
		hasWidth bool
	)
	for x, battr = range listAttribute {
		if x == 0 {
			var alt = bytes.TrimSpace(listAttribute[0])
			if len(alt) == 0 {
				var dot = bytes.IndexByte(link, '.')
				if dot > 0 {
					alt = link[:dot]
				}
			}
			el.Attrs[attrNameAlt] = string(alt)
			continue
		}
		if x == 1 {
			if ascii.IsDigits(listAttribute[1]) {
				el.Attrs[attrNameWidth] = string(listAttribute[1])
				hasWidth = true
				continue
			}
		}
		if hasWidth && x == 2 {
			if ascii.IsDigits(listAttribute[2]) {
				el.Attrs[attrNameHeight] = string(listAttribute[2])
			}
		}

		var attr = string(battr)
		var kv = strings.SplitN(attr, `=`, 2)
		if len(kv) != 2 {
			continue
		}
		var key = strings.TrimSpace(kv[0])
		var val = strings.Trim(kv[1], `"`)

		switch key {
		case attrNameFloat, attrNameAlign, attrNameRole:
			if val == `center` {
				val = `text-center`
			}
			el.addRole(val)

		case attrNameLink:
			val = string(applySubstitutions(doc, []byte(val)))
			el.Attrs[key] = val

		default:
			el.Attrs[key] = val
		}
	}
	return true
}

func parseInlineImage(doc *Document, content []byte) (el *element, n int) {
	// If the next character is ':' (as in block "image::") mark it as
	// invalid inline image, since this is block image that has been
	// parsed but invalid (probably missing '[]').
	if content[0] == ':' {
		return nil, 0
	}

	_, n = indexByteUnescape(content, ']')
	if n < 0 {
		return nil, 0
	}

	var lineImage = content[:n+1]
	el = &element{
		elementAttribute: elementAttribute{
			Attrs: make(map[string]string),
		},
		kind: elKindInlineImage,
	}
	if el.parseBlockImage(doc, lineImage) {
		return el, n + 2
	}
	return nil, 0
}

func htmlWriteInlineImage(el *element, out io.Writer) {
	var (
		classes = strings.TrimSpace(`image ` + el.htmlClasses())

		link     string
		withLink bool
	)

	fmt.Fprintf(out, `<span class=%q>`, classes)
	link, withLink = el.Attrs[attrNameLink]
	if withLink {
		fmt.Fprintf(out, `<a class=%q href=%q>`, attrValueImage, link)
	}

	var (
		src = el.Attrs[attrNameSrc]
		alt = el.Attrs[attrNameAlt]

		width  string
		height string
		ok     bool
	)

	width, ok = el.Attrs[attrNameWidth]
	if ok {
		width = fmt.Sprintf(` width="%s"`, width)
	}
	height, ok = el.Attrs[attrNameHeight]
	if ok {
		height = fmt.Sprintf(` height="%s"`, height)
	}

	fmt.Fprintf(out, `<img src=%q alt=%q%s%s>`, src, alt, width, height)

	if withLink {
		fmt.Fprint(out, `</a>`)
	}

	fmt.Fprint(out, `</span>`)
}
