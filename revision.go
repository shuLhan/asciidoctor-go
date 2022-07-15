// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"strings"
)

// Revision contains the document version, date, and remark.
type Revision struct {
	Number string
	Date   string
	Remark string
}

// parseRevision parse document revision in the following format,
//
//	DOC_REVISION     = DOC_REV_VERSION [ "," DOC_REV_DATE ]
//
//	DOC_REV_VERSION  = "v" 1*DIGIT "." 1*DIGIT "." 1*DIGIT
//
//	DOC_REV_DATE     = 1*2DIGIT WSP 3*ALPHA WSP 4*DIGIT
func parseRevision(raw string) Revision {
	var (
		rev   Revision
		x     int
		lastc byte
	)
	if len(raw) == 0 {
		return rev
	}
	if raw[0] == 'v' {
		for x = 0; x < len(raw); x++ {
			if raw[x] == ',' || raw[x] == ':' {
				lastc = raw[x]
				break
			}
		}
		rev.Number = strings.TrimSpace(raw[1:x])
		if x == len(raw) {
			return rev
		}
		raw = raw[x:]
		x = 0
	}
	if len(raw) > 0 && raw[0] != ':' {
		// Parse date, check for option remark that start with ':'.
		if lastc != 0 {
			raw = raw[1:]
		}
		for x = 0; x < len(raw); x++ {
			if raw[x] == ':' {
				break
			}
		}
		rev.Date = strings.TrimSpace(raw[:x])
		if x == len(raw) {
			return rev
		}
		raw = raw[x:]
		x = 0
	}
	if len(raw) > 0 {
		// We have left over as remark.
		rev.Remark = strings.TrimSpace(raw[x+1:])
	}
	return rev
}
