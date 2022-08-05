// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestInlineParser(t *testing.T) {
	var (
		_testDoc = &Document{
			Attributes: map[string]string{
				`x`: `https://kilabit.info`,
			},

			anchors: map[string]*anchor{
				`x`: &anchor{
					label: `X y`,
				},
			},
			titleID: map[string]string{
				`X y`: `x`,
			},
		}

		buf       bytes.Buffer
		tdata     *test.Data
		container *element
		err       error
		name      string
		lineNum   string
		vbytes    []byte
		exps      [][]byte
		lines     [][]byte
		x         int
	)

	tdata, err = test.LoadData(`testdata/inline_parser/inline_parser_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	for name, vbytes = range tdata.Input {
		t.Run(name, func(t *testing.T) {
			lines = bytes.Split(vbytes, []byte("\n"))
			exps = bytes.Split(tdata.Output[name], []byte("\n"))

			for x, vbytes = range lines {
				buf.Reset()
				container = parseInlineMarkup(_testDoc, vbytes)
				container.toHTML(_testDoc, &buf)

				lineNum = fmt.Sprintf("#%d", x)
				test.Assert(t, lineNum, string(exps[x]), buf.String())
			}
		})
	}
}

func TestInlineParser_parseInlineID_isForToC(t *testing.T) {
	type testCase struct {
		content string
		exp     string
	}

	var _testDoc = &Document{
		anchors:  make(map[string]*anchor),
		titleID:  make(map[string]string),
		isForToC: true,
	}

	var cases = []testCase{{
		content: `[[A]] B`,
		exp:     ` B`,
	}}

	var (
		buf       bytes.Buffer
		c         testCase
		container *element
		got       string
	)

	for _, c = range cases {
		buf.Reset()

		container = parseInlineMarkup(_testDoc, []byte(c.content))
		container.toHTML(_testDoc, &buf)

		got = buf.String()
		test.Assert(t, c.content, c.exp, got)
	}
}
