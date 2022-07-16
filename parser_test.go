// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsValidID(t *testing.T) {
	type testCase struct {
		id  string
		exp bool
	}

	var cases = []testCase{{
		id:  "a",
		exp: true,
	}, {
		id: "1",
	}}

	var (
		c   testCase
		got bool
	)

	for _, c = range cases {
		got = isValidID([]byte(c.id))
		test.Assert(t, c.id, c.exp, got)
	}
}
