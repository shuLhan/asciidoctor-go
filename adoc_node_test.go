// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestAdocNode_parseListDescription(t *testing.T) {
	cases := []struct {
		line       string
		expLevel   int
		expRawTerm string
		expRaw     string
	}{{
		line:       "CPU::",
		expLevel:   0,
		expRawTerm: "CPU",
	}}

	for _, c := range cases {
		node := &adocNode{}
		node.parseListDescription(c.line)

		test.Assert(t, "adocNode.Level", c.expLevel, node.level, true)
		test.Assert(t, "adocNode.rawLabel", c.expRawTerm, node.rawLabel.String(), true)
		test.Assert(t, "adocNode.raw", c.expRaw, string(node.raw), true)
	}
}
