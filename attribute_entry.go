// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "strings"

// AttributeEntry contains the mapping of global attribute keys in the headers
// with its value.
type AttributeEntry map[string]string

func newAttributeEntry() AttributeEntry {
	return AttributeEntry{
		MetaNameGenerator:    `asciidoctor-go ` + Version,
		metaNameSectIDs:      ``,
		metaNameShowTitle:    ``,
		metaNameTableCaption: ``,
		metaNameVersionLabel: ``,
	}
}

func (entry *AttributeEntry) apply(key, val string) {
	if key[0] == '!' {
		delete(*entry, strings.TrimSpace(key[1:]))
	} else if key[len(key)-1] == '!' {
		delete(*entry, strings.TrimSpace(key[:len(key)-1]))
	} else {
		(*entry)[key] = val
	}
}
