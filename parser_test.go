// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsValidID(t *testing.T) {
	cases := []struct {
		id  string
		exp bool
	}{{
		id:  "a",
		exp: true,
	}, {
		id: "1",
	}}

	for _, c := range cases {
		got := isValidID([]byte(c.id))
		test.Assert(t, c.id, c.exp, got)
	}
}
