// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseRevision(t *testing.T) {
	cases := []struct {
		raw string
		exp Revision
	}{{
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

	for _, c := range cases {
		got := parseRevision(c.raw)
		test.Assert(t, "Revision", c.exp, got, true)
	}
}
