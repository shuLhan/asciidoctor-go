// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "strings"

//
// AttributeEntry contains the mapping of global attribute keys in the headers
// with its value.
//
type AttributeEntry map[string]string

func newAttributeEntry() AttributeEntry {
	return AttributeEntry{
		metaNameSectIDs:      "",
		metaNameShowTitle:    "",
		metaNameVersionLabel: "",
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
