// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"strings"

	libstrings "github.com/shuLhan/share/lib/strings"
)

type attributeClass []string

func (aclass *attributeClass) add(c string) {
	(*aclass) = libstrings.AppendUniq(*aclass, c)
}

func (aclass *attributeClass) replace(old, new string) {
	var (
		v string
		x int
	)

	for x, v = range *aclass {
		if v == old {
			(*aclass)[x] = new
			return
		}
	}
	// Add class if not found.
	aclass.add(new)
}

// String concat all the attribute class into string separated by single
// space.
func (aclass attributeClass) String() string {
	return strings.Join(aclass, " ")
}
