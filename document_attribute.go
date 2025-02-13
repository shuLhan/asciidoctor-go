// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import (
	"fmt"
	"strconv"
	"strings"
)

// List of document attribute.
const (
	DocAttrAuthor      = `author`       // May contain the first author full name only.
	DocAttrAuthorNames = `author_names` // List of author full names, separated by comma.
	DocAttrDescription = `description`
	DocAttrGenerator   = `generator`
	DocAttrKeywords    = `keywords`

	docAttrAuthorInitials  = `authorinitials`
	docAttrDocTitle        = `doctitle`
	docAttrEmail           = attrValueEmail
	docAttrFirstName       = `firstname`
	docAttrIDPrefix        = `idprefix`
	docAttrIDSeparator     = `idseparator`
	docAttrLastName        = `lastname`
	docAttrLastUpdateLabel = `last-update-label`
	docAttrLastUpdateValue = `last-update-value`
	docAttrLevelOffset     = `leveloffset`
	docAttrMiddleName      = `middlename`
	docAttrNoFooter        = `nofooter`
	docAttrNoHeader        = `noheader`
	docAttrNoHeaderFooter  = `no-header-footer`
	docAttrNoTitle         = `notitle`
	docAttrRevDate         = `revdate`
	docAttrRevNumber       = `revnumber`
	docAttrRevRemark       = `revremark`
	docAttrSectAnchors     = `sectanchors`
	docAttrSectIDs         = `sectids`
	docAttrSectLinks       = `sectlinks`
	docAttrSectNumLevel    = `sectnumlevels`
	docAttrSectNums        = `sectnums`
	docAttrShowTitle       = `showtitle`
	docAttrStylesheet      = `stylesheet`
	docAttrTOC             = `toc`
	docAttrTOCLevels       = `toclevels`
	docAttrTOCTitle        = `toc-title`
	docAttrTableCaption    = `table-caption`
	docAttrTitle           = attrNameTitle
	docAttrTitleSeparator  = `title-separator`
	docAttrVersionLabel    = `version-label`
)

// List of possible document attribute value.
const (
	docAttrValueAuto     = `auto`
	docAttrValueMacro    = `macro`
	docAttrValuePreamble = `preamble`
	docAttrValueLeft     = `left`
	docAttrValueRight    = `right`
)

// DocumentAttribute contains the mapping of global attribute keys in the
// headers with its value.
type DocumentAttribute struct {
	Entry       map[string]string
	LevelOffset int
}

func newDocumentAttribute() DocumentAttribute {
	return DocumentAttribute{
		Entry: map[string]string{
			DocAttrGenerator:       `asciidoctor-go ` + Version,
			docAttrLastUpdateLabel: `Last updated`,
			docAttrLastUpdateValue: ``,
			docAttrSectIDs:         ``,
			docAttrShowTitle:       ``,
			docAttrStylesheet:      ``,
			docAttrTableCaption:    ``,
			docAttrVersionLabel:    ``,
		},
	}
}

func (docAttr *DocumentAttribute) apply(key, val string) (err error) {
	if key[0] == '!' {
		key = strings.TrimSpace(key[1:])
		delete(docAttr.Entry, key)
		return nil
	}
	var n = len(key)
	if key[n-1] == '!' {
		key = strings.TrimSpace(key[:n-1])
		delete(docAttr.Entry, key)
		return nil
	}

	if key == docAttrLevelOffset {
		var offset int64
		offset, err = strconv.ParseInt(val, 10, 32)
		if err != nil {
			return fmt.Errorf(`DocumentAttribute: %s: invalid value %q`, key, val)
		}
		if val[0] == '+' || val[0] == '-' {
			docAttr.LevelOffset += int(offset)
			goto valid
		}
		docAttr.LevelOffset = int(offset)
	}

valid:
	docAttr.Entry[key] = val
	return nil
}
