// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
	if level == sec.curr {
		sec.nums[level]++
	} else if level > sec.curr {
		// Check if the section level out of sequence.
		if level > sec.curr+1 {
			level = sec.curr + 1
		}
		sec.nums[level] = 1
		sec.curr = level
	} else {
		for x := sec.curr; x > level; x-- {
			sec.nums[x] = 0
		}
		sec.nums[level]++
		sec.curr = level
	}
	clone := *sec
	return &clone
}

func (sec *sectionCounters) String() string {
	var (
		sb strings.Builder
	)

	for x := 1; x < 6; x++ {
		if sec.nums[x] == 0 {
			break
		}
		fmt.Fprintf(&sb, "%d.", sec.nums[x])
	}
	sb.WriteByte(' ')
	return sb.String()
}
