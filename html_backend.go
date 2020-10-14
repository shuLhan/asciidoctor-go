// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "strings"

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
