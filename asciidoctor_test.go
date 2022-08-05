// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	outputCallHtmlWriteHeader = "htmlWriteHeader"
	outputCallToHTML          = "ToHTML"
	outputCallToHTMLBody      = "ToHTMLBody"
)

func TestData(t *testing.T) {
	var (
		listTData    []*test.Data
		tdata        *test.Data
		inputCall    string
		outputCall   string
		inputName    string
		subtestName  string
		inputContent []byte
		err          error
	)

	listTData, err = test.LoadDataDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTData {
		inputCall = tdata.Flag["input_call"]
		outputCall = tdata.Flag["output_call"]

		for inputName, inputContent = range tdata.Input {
			subtestName = tdata.Name + "/" + inputName

			t.Run(subtestName, func(t *testing.T) {
				var (
					doc  *Document
					bbuf bytes.Buffer
					exp  []byte
					got  []byte
				)

				exp = tdata.Output[inputName]

				bbuf.Reset()

				switch inputCall {
				default:
					doc = Parse(inputContent)
				}

				switch outputCall {
				case outputCallHtmlWriteHeader:
					htmlWriteHeader(doc, &bbuf)
				case outputCallToHTML:
					err = doc.ToHTML(&bbuf)
				case outputCallToHTMLBody:
					err = doc.ToHTMLBody(&bbuf)
				default:
					err = doc.ToHTMLEmbedded(&bbuf)
				}
				if err != nil {
					got = []byte(err.Error())
				} else {
					got = bbuf.Bytes()
				}

				test.Assert(t, subtestName, string(exp), string(got))
			})
		}
	}
}
