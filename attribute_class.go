// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"strings"

	libstrings "github.com/shuLhan/share/lib/strings"
)

type attributeClass []string

func (aclass *attributeClass) add(c string) {
	(*aclass) = libstrings.AppendUniq(*aclass, c)
}

//
// String concat all the attribute class into string separated by single
// space.
//
func (aclass attributeClass) String() string {
	return strings.Join(aclass, " ")
}
