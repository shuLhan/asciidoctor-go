// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "strings"

type attributeClass map[string]struct{}

func (aclass *attributeClass) add(c string) {
	if *aclass == nil {
		*aclass = attributeClass{}
	}
	(*aclass)[c] = struct{}{}
}

//
// String concat all the attribute class keys into string separated by single
// space.
//
func (aclass attributeClass) String() string {
	var (
		sb strings.Builder
		x  int
	)
	for k := range aclass {
		if x > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		x++
	}
	return sb.String()
}