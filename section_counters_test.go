// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSectionCounters(t *testing.T) {
	sec := &sectionCounters{}

	cases := []struct {
		level     int
		exp       *sectionCounters
		expString string
	}{{
		level: 2,
		exp: &sectionCounters{
			nums: [6]byte{0, 1, 0, 0, 0, 0},
			curr: 1,
		},
		expString: "1. ",
	}, {
		level: 1,
		exp: &sectionCounters{
			nums: [6]byte{0, 2, 0, 0, 0, 0},
			curr: 1,
		},
		expString: "2. ",
	}, {
		level: 2,
		exp: &sectionCounters{
			nums: [6]byte{0, 2, 1, 0, 0, 0},
			curr: 2,
		},
		expString: "2.1. ",
	}, {
		level: 3,
		exp: &sectionCounters{
			nums: [6]byte{0, 2, 1, 1, 0, 0},
			curr: 3,
		},
		expString: "2.1.1. ",
	}, {
		level: 2,
		exp: &sectionCounters{
			nums: [6]byte{0, 2, 2, 0, 0, 0},
			curr: 2,
		},
		expString: "2.2. ",
	}, {
		level: 1,
		exp: &sectionCounters{
			nums: [6]byte{0, 3, 0, 0, 0, 0},
			curr: 1,
		},
		expString: "3. ",
	}}

	for _, c := range cases {
		got := sec.set(c.level)
		gotString := got.String()
		test.Assert(t, "set", c.exp, got)
		test.Assert(t, "String", c.expString, gotString)
	}
}
