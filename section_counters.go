// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"fmt"
	"strings"
)

type sectionCounters struct {
	// index 0 represent counter for level 0,
	// index 1 represent coutner for level 1, and so on.
	nums [6]byte
	curr int
}

func (sec *sectionCounters) set(level int) *sectionCounters {
	switch {
	case level == sec.curr:
		sec.nums[level]++
	case level > sec.curr:
		// Check if the section level out of sequence.
		if level > sec.curr+1 {
			level = sec.curr + 1
		}
		sec.nums[level] = 1
		sec.curr = level
	default:
		var x int
		for x = sec.curr; x > level; x-- {
			sec.nums[x] = 0
		}
		sec.nums[level]++
		sec.curr = level
	}
	var clone = *sec
	return &clone
}

func (sec *sectionCounters) String() string {
	var (
		sb strings.Builder
		x  int
	)

	for x = 1; x < 6; x++ {
		if sec.nums[x] == 0 {
			break
		}
		fmt.Fprintf(&sb, `%d.`, sec.nums[x])
	}
	sb.WriteByte(' ')
	return sb.String()
}
