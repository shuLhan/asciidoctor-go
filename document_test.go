// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
