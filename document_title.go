// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import "fmt"

type DocumentTitle struct {
	Main string
	Sub  string
	el   *element
	raw  string
	sep  byte
}

//
// String return the combination of main and subtitle separated by colon or
// meta "title-separator" value.
//
func (docTitle *DocumentTitle) String() string {
	if len(docTitle.Sub) > 0 {
		return fmt.Sprintf("%s%c %s", docTitle.Main, docTitle.sep, docTitle.Sub)
	}
	return docTitle.Main
}
