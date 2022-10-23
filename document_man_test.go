// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestDocument_ToMan(t *testing.T) {
	type testCase struct {
		name     string
		fileAdoc string
		fileExp  string
		fileGot  string
	}

	var (
		expDate = `2022-12-15`
		cases   = []testCase{{
			name:     `section`,
			fileAdoc: `testdata/man/section.adoc`,
			fileExp:  `testdata/man/section.exp.7`,
			fileGot:  `testdata/man/section.got.7`,
		}, {
			name:     `inline format`,
			fileAdoc: `testdata/man/inline_format.adoc`,
			fileExp:  `testdata/man/inline_format.exp.7`,
			fileGot:  `testdata/man/inline_format.got.7`,
		}, {
			name:     `list`,
			fileAdoc: `testdata/man/list.adoc`,
			fileExp:  `testdata/man/list.exp.7`,
			fileGot:  `testdata/man/list.got.7`,
		}, {
			name:     `macro`,
			fileAdoc: `testdata/man/macro.adoc`,
			fileExp:  `testdata/man/macro.exp.7`,
			fileGot:  `testdata/man/macro.got.7`,
		}}

		c       testCase
		got     bytes.Buffer
		gotFile *os.File
		doc     *Document
		exp     []byte
		err     error
	)

	for _, c = range cases {
		exp, err = os.ReadFile(c.fileExp)
		if err != nil {
			t.Fatal(err)
		}

		doc, err = Open(c.fileAdoc)
		if err != nil {
			t.Fatal(err)
		}

		gotFile, err = os.Create(c.fileGot)
		if err != nil {
			t.Fatal(err)
		}

		// Set the asciidoctor version to current Asciidoctor version.
		doc.Attributes[MetaNameGenerator] = `Asciidoctor 2.0.18`
		doc.Revision.Date = expDate

		got.Reset()

		err = doc.ToMan(&got)
		if err != nil {
			t.Fatal(err)
		}

		_, err = gotFile.Write(got.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, `ToMan `+c.name, string(exp), got.String())
	}
}
