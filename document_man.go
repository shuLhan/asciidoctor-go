// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"io"
	"strings"
	"time"
)

// List of document attributes for manpage.
const (
	MetanameManLinkstyle = `man-linkstyle`
	MetanameManManual    = `manmanual`
	MetanameManSource    = `mansource`
	MetanameManTitle     = `mantitle`
	MetanameManVolnum    = `manvolnum`
)

const (
	defManVolnum    = `7`
	defManLinkstyle = `blue R < >`
)

// initMan initialize the default values for man document type.
func (doc *Document) initMan() {
	var (
		manTitle, manVolnum = parseManTitle(doc.Title.Main)
		manTitleUp          = manTitle
		timeNow             = time.Now()
		dateNow             = timeNow.Format(`2006-01-02`)

		ok bool
	)

	manTitleUp = strings.ToUpper(manTitleUp)

	if len(manVolnum) == 0 {
		manVolnum = defManVolnum
	}

	if len(doc.Revision.Date) == 0 {
		doc.Revision.Date = dateNow
	}

	_, ok = doc.Attributes[MetanameManLinkstyle]
	if !ok {
		doc.Attributes[MetanameManLinkstyle] = defManLinkstyle
	}

	_, ok = doc.Attributes[MetanameManManual]
	if !ok {
		doc.Attributes[MetanameManManual] = manTitleUp
	}

	_, ok = doc.Attributes[MetanameManSource]
	if !ok {
		doc.Attributes[MetanameManSource] = manTitleUp
	}

	_, ok = doc.Attributes[MetanameManTitle]
	if !ok {
		doc.Attributes[MetanameManTitle] = manTitle
	}

	_, ok = doc.Attributes[MetanameManVolnum]
	if !ok {
		doc.Attributes[MetanameManVolnum] = manVolnum
	}
}

// ToMan convert the Document into UNIX manual page (ROFF document format).
func (doc *Document) ToMan(out io.Writer) (err error) {
	doc.initMan()

	roffWriteHeader(out, doc)
	roffWriteBody(out, doc)
	roffWriteFooter(out, doc)

	return nil
}

// parseManTitle parse the title and section from text in the following
// format: title(volnum).
func parseManTitle(text string) (title, volnum string) {
	var (
		idx = strings.IndexByte(text, '(')
	)
	if idx < 0 {
		title = text
	} else {
		title = text[:idx]

		if idx < len(text) {
			if text[idx+1] >= '1' && text[idx+1] <= '8' {
				volnum = string(text[idx+1])
			}
		}
	}
	return title, volnum
}
