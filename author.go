// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"strings"
)

// Author of document.
type Author struct {
	FirstName  string
	MiddleName string
	LastName   string
	Initials   string
	Email      string
}

// parseAuthor parse raw author into object.
func parseAuthor(raw string) (author *Author) {
	var (
		names   []string
		idx     int
		lastIdx int
	)

	author = &Author{}

	raw = strings.TrimSpace(raw)
	if raw[len(raw)-1] == '>' {
		idx = strings.IndexByte(raw, '<')
		if idx > 0 {
			author.Email = raw[idx+1 : len(raw)-1]
			raw = strings.TrimSpace(raw[:idx])
		}
	}

	names = strings.Split(raw, " ")
	if len(names) == 0 {
		return
	}

	var initials bytes.Buffer
	author.FirstName = strings.ReplaceAll(names[0], "_", " ")
	initials.WriteByte(author.FirstName[0])

	if len(names) >= 2 {
		lastIdx = len(names) - 1
		author.LastName = strings.ReplaceAll(names[lastIdx], "_", " ")

		author.MiddleName = strings.ReplaceAll(
			strings.Join(names[1:lastIdx], " "), "_", " ",
		)

		if len(author.MiddleName) > 0 {
			initials.WriteByte(author.MiddleName[0])
		}
		initials.WriteByte(author.LastName[0])
	}

	author.Initials = initials.String()

	return author
}

// FullName return the concatenation of author first, middle, and last name.
func (author *Author) FullName() string {
	var sb strings.Builder

	sb.WriteString(author.FirstName)
	if len(author.MiddleName) > 0 {
		sb.WriteString(" " + author.MiddleName)
	}
	if len(author.LastName) > 0 {
		sb.WriteString(" " + author.LastName)
	}
	return sb.String()
}
