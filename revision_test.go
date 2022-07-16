// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseRevision(t *testing.T) {
	type testCase struct {
		raw string
		exp Revision
	}

	var cases = []testCase{{
		raw: "v1",
		exp: Revision{
			Number: "1",
		},
	}, {
		raw: "15 Nov, 2020",
		exp: Revision{
			Date: "15 Nov, 2020",
		},
	}, {
		raw: ":remark",
		exp: Revision{
			Remark: "remark",
		},
	}, {
		raw: "v1, 15 Nov, 2020",
		exp: Revision{
			Number: "1",
			Date:   "15 Nov, 2020",
		},
	}, {
		raw: "v1: remark",
		exp: Revision{
			Number: "1",
			Remark: "remark",
		},
	}, {
		raw: "15 Nov, 2020: remark",
		exp: Revision{
			Date:   "15 Nov, 2020",
			Remark: "remark",
		},
	}, {
		raw: "v1, 15 Nov: remark",
		exp: Revision{
			Number: "1",
			Date:   "15 Nov",
			Remark: "remark",
		},
	}}

	var (
		c   testCase
		got Revision
	)

	for _, c = range cases {
		got = parseRevision(c.raw)
		test.Assert(t, "Revision", c.exp, got)
	}
}
