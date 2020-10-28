// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"log"
	"os"
	"testing"
	"text/template"
)

var (
	_testDoc  *Document
	_testTmpl *template.Template
)

func TestMain(m *testing.M) {
	var err error

	_testDoc = &Document{}
	_testTmpl, err = _testDoc.createHTMLTemplate()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
