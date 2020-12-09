// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

type elementInclude struct {
	fpath   string
	attrs   elementAttribute
	content []byte
}

func parseInclude(doc *Document, line []byte) (el *elementInclude) {
	var err error

	if !bytes.HasPrefix(line, []byte(prefixInclude)) {
		return nil
	}

	el = &elementInclude{}
	line = line[len(prefixInclude):]

	path, start := indexByteUnescape(line, '[')
	if start == -1 {
		return nil
	}
	_, end := indexByteUnescape(line[start:], ']')
	if end == -1 {
		return nil
	}

	el.attrs.parseElementAttribute(string(line[start : start+end+1]))

	path = applySubstitutions(doc, path)
	el.fpath = filepath.Join(filepath.Dir(doc.fpath), string(path))

	fmt.Printf("parseInclude: path:%s fpath:%s\n", path, el.fpath)

	el.content, err = ioutil.ReadFile(el.fpath)
	if err != nil {
		log.Printf("parseInclude %q: %s", doc.file, err)
		return nil
	}

	return el
}
