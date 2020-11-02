// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "strings"

type attributeEntry struct {
	v map[string]string
}

func newAttributeEntry() *attributeEntry {
	return &attributeEntry{
		v: make(map[string]string),
	}
}

func (entry *attributeEntry) apply(key, val string) {
	if key[0] == '!' {
		delete(entry.v, strings.TrimSpace(key[1:]))
	} else if key[len(key)-1] == '!' {
		delete(entry.v, strings.TrimSpace(key[:len(key)-1]))
	} else {
		entry.v[key] = val
	}
}
