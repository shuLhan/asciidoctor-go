// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"bytes"
	"log"
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

	path = applySubstitutions(doc, path)
	el.fpath = filepath.Join(filepath.Dir(doc.fpath), string(path))

	el.content, err = os.ReadFile(el.fpath)
	if err != nil {
		log.Printf("parseInclude %q: %s", doc.file, err)
		return nil
	}

	return el
}
