// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "fmt"

type DocumentTitle struct {
	Main string
	Sub  string
	el   *element
	raw  string
	sep  byte
}

// String return the combination of main and subtitle separated by colon or
// meta "title-separator" value.
func (docTitle *DocumentTitle) String() string {
	if len(docTitle.Sub) > 0 {
		return fmt.Sprintf("%s%c %s", docTitle.Main, docTitle.sep, docTitle.Sub)
	}
	return docTitle.Main
}
