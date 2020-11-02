// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

const (
	classNameArticle      = "article"
	classNameListingBlock = "listingblock"
	classNameLiteralBlock = "literalblock"
	classNameToc          = "toc"
	classNameToc2         = "toc2"
	classNameTocLeft      = "toc-left"
	classNameTocRight     = "toc-right"
	classNameUlist        = "ulist"
)

const (
	htmlLessthanSymbol    = "&lt;"
	htmlGreaterthanSymbol = "&gt;"
	htmlAmpersandSymbol   = "&amp;"
)

func htmlSubstituteSpecialChars(in string) (out string) {
	var (
		isEscaped bool
		sb        strings.Builder
	)
	sb.Grow(len(in))

	for _, c := range in {
		if isEscaped {
			if c == '\\' || c == '<' || c == '>' || c == '&' {
				sb.WriteRune(c)
			} else {
				sb.WriteRune('\\')
				sb.WriteRune(c)
			}
			isEscaped = false
			continue
		}
		switch c {
		case '\\':
			isEscaped = true
		case '<':
			sb.WriteString(htmlLessthanSymbol)
		case '>':
			sb.WriteString(htmlGreaterthanSymbol)
		case '&':
			sb.WriteString(htmlAmpersandSymbol)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func (doc *Document) htmlGenerateTOC(
	node *adocNode, tmpl *template.Template, out io.Writer, level int,
) (err error) {
	if level >= doc.TOCLevel {
		return nil
	}

	var sectClass string

	switch node.kind {
	case nodeKindSectionL1:
		sectClass = "sectlevel1"
	case nodeKindSectionL2:
		sectClass = "sectlevel2"
	case nodeKindSectionL3:
		sectClass = "sectlevel3"
	case nodeKindSectionL4:
		sectClass = "sectlevel4"
	case nodeKindSectionL5:
		sectClass = "sectlevel5"
	}

	if len(sectClass) > 0 {
		if level < node.level {
			_, err = fmt.Fprintf(out, "\n<ul class=\"%s\">", sectClass)
		} else if level > node.level {
			n := level
			for n > node.level {
				_, err = out.Write([]byte("\n</ul>"))
				n--
			}
		}

		_, err = fmt.Fprintf(out, "\n<li><a href=\"#%s\">", node.ID)
		if err != nil {
			return fmt.Errorf("htmlGenerateTOC: %w", err)
		}

		if node.sectnums != nil {
			_, err = out.Write([]byte(node.sectnums.String()))
			if err != nil {
				return fmt.Errorf("htmlGenerateTOC: %w", err)
			}
		}

		err = node.title.toHTML(doc, tmpl, out, true)
		if err != nil {
			return fmt.Errorf("htmlGenerateTOC: %w", err)
		}

		_, err = out.Write([]byte("</a>"))
		if err != nil {
			return fmt.Errorf("htmlGenerateTOC: %w", err)
		}
	}

	if node.child != nil {
		err = doc.htmlGenerateTOC(node.child, tmpl, out, node.level)
		if err != nil {
			return err
		}
	}
	if len(sectClass) > 0 {
		_, err = out.Write([]byte("</li>"))
		if err != nil {
			return fmt.Errorf("htmlGenerateTOC: %w", err)
		}
	}
	if node.next != nil {
		err = doc.htmlGenerateTOC(node.next, tmpl, out, node.level)
		if err != nil {
			return err
		}
	}

	if len(sectClass) > 0 && level < node.level {
		_, err = out.Write([]byte("\n</ul>\n"))
		if err != nil {
			return fmt.Errorf("htmlGenerateTOC: %w", err)
		}
	}

	return nil
}
