// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "strings"

// DocumentAttribute contains the mapping of global attribute keys in the
// headers with its value.
type DocumentAttribute map[string]string

func newDocumentAttribute() DocumentAttribute {
	return DocumentAttribute{
		MetaNameGenerator:       `asciidoctor-go ` + Version,
		metaNameLastUpdateLabel: `Last updated`,
		metaNameLastUpdateValue: ``,
		metaNameSectIDs:         ``,
		metaNameShowTitle:       ``,
		metaNameTableCaption:    ``,
		metaNameVersionLabel:    ``,
	}
}

func (entry *DocumentAttribute) apply(key, val string) {
	switch {
	case key[0] == '!':
		delete(*entry, strings.TrimSpace(key[1:]))
	case key[len(key)-1] == '!':
		delete(*entry, strings.TrimSpace(key[:len(key)-1]))
	default:
		(*entry)[key] = val
	}
}
