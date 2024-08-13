// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "strings"

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
type DocumentAttribute map[string]string

func newDocumentAttribute() DocumentAttribute {
	return DocumentAttribute{
		DocAttrGenerator:       `asciidoctor-go ` + Version,
		docAttrLastUpdateLabel: `Last updated`,
		docAttrLastUpdateValue: ``,
		docAttrSectIDs:         ``,
		docAttrShowTitle:       ``,
		docAttrTableCaption:    ``,
		docAttrVersionLabel:    ``,
	}
}

func (entry *DocumentAttribute) apply(key, val string) {
	switch {
	case key[0] == '!':
		delete(*entry, strings.TrimSpace(key[1:]))
	case key[len(key)-1] == '!':
		delete(*entry, strings.TrimSpace(key[:len(key)-1]))
	default:
		(*entry)[key] = val
	}
}