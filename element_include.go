// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"os"
	"path/filepath"
)

type elementInclude struct {
	fpath   string
	content []byte
	attrs   elementAttribute
}

func parseInclude(doc *Document, line []byte) (el *elementInclude) {
	var (
		path  []byte
		start int
		end   int
		err   error
	)

	if !bytes.HasPrefix(line, []byte(prefixInclude)) {
		return nil
	}

	el = &elementInclude{}
	line = line[len(prefixInclude):]

	path, start = indexByteUnescape(line, '[')
	if start == -1 {
		return nil
	}

	_, end = indexByteUnescape(line[start:], ']')
	if end == -1 {
		return nil
	}

	el.attrs.parseElementAttribute(line[start : start+end+1])

	var newPath = applySubstitutions(doc, path)
	if bytes.Contains(path, []byte(docAttrDocdir)) {
		el.fpath = string(newPath)
	} else {
		el.fpath = filepath.Join(doc.docdir, string(newPath))
	}
	el.content, err = os.ReadFile(el.fpath)
	if err != nil {
		return nil
	}

	return el
}
