// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	outputCallHTMLWriteHeader = `htmlWriteHeader`
	outputCallToHTML          = `ToHTML`
	outputCallToHTMLBody      = `ToHTMLBody`
)

func TestData(t *testing.T) {
	var (
		listTData    []*test.Data
		tdata        *test.Data
		bbuf         bytes.Buffer
		outputCall   string
		inputName    string
		subtestName  string
		inputContent []byte
		exp          []byte
		got          []byte
		err          error
	)

	listTData, err = test.LoadDataDir(`testdata`)
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTData {
		outputCall = tdata.Flag[`output_call`]

		for inputName, inputContent = range tdata.Input {
			subtestName = tdata.Name + `/` + inputName

			t.Run(subtestName, func(t *testing.T) {
				bbuf.Reset()

				var doc = Parse(inputContent)

				switch outputCall {
				case outputCallHTMLWriteHeader:
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

				exp = tdata.Output[inputName]
				test.Assert(t, subtestName, string(exp), string(got))
			})
		}
	}
}
