// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSectionCounters(t *testing.T) {
	type testCase struct {
		exp       *sectionCounters
		expString string
		level     int
	}

	var cases = []testCase{{
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

	var (
		sec = &sectionCounters{}

		got       *sectionCounters
		gotString string
		c         testCase
	)

	for _, c = range cases {
		got = sec.set(c.level)
		gotString = got.String()
		test.Assert(t, "set", c.exp, got)
		test.Assert(t, "String", c.expString, gotString)
	}
}
