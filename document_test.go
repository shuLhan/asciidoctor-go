// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	doc, err := Open("testdata/test.adoc")
	if err != nil {
		t.Fatal(err)
	}

	fout, err := os.OpenFile("testdata/got.test.html",
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = doc.ToHTML(fout)
	if err != nil {
		t.Fatal(err)
	}
}
