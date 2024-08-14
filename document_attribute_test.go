// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestDocumentAttributeApply(t *testing.T) {
	type testCase struct {
		desc     string
		key      string
		val      string
		expError string
		exp      DocumentAttribute
	}

	var docAttr = DocumentAttribute{
		Entry: map[string]string{
			`key1`: ``,
			`key2`: ``,
		},
	}

	var listCase = []testCase{{
		key: `key3`,
		exp: docAttr,
	}, {
		desc: `prefix negation`,
		key:  `!key1`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key2`: ``,
				`key3`: ``,
			},
		},
	}, {
		desc: `suffix negation`,
		key:  `key2!`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`: ``,
			},
		},
	}, {
		desc: `leveloffset +`,
		key:  docAttrLevelOffset,
		val:  `+2`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `+2`,
			},
			LevelOffset: 2,
		},
	}, {
		desc: `leveloffset -`,
		key:  docAttrLevelOffset,
		val:  `-2`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `-2`,
			},
			LevelOffset: 0,
		},
	}, {
		desc: `leveloffset`,
		key:  docAttrLevelOffset,
		val:  `1`,
		exp: DocumentAttribute{
			Entry: map[string]string{
				`key3`:        ``,
				`leveloffset`: `1`,
			},
			LevelOffset: 1,
		},
	}, {
		desc:     `leveloffset: invalid`,
		key:      docAttrLevelOffset,
		val:      `*1`,
		expError: `DocumentAttribute: leveloffset: invalid value "*1"`,
	}}

	var (
		tc  testCase
		err error
	)
	for _, tc = range listCase {
		err = docAttr.apply(tc.key, tc.val)
		if err != nil {
			test.Assert(t, `apply: `+tc.desc, tc.expError, err.Error())
			continue
		}
		test.Assert(t, `apply: `+tc.desc, tc.exp, docAttr)
	}
}
