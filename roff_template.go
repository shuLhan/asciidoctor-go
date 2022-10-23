// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

const (
	_roffTmplBlockLiteralBegin = `.sp
.if n .RS 4
.nf
.fam C
`
	_roffTmplBlockLiteralEnd = `.fam
.fi
.if n .RE
`
)

// List of roff template for ordered list.
const (
	// roff macro for list item with parameter its level.
	_roffTmplListOrderedItemBegin = `.RS 4
.ie n \{\
\h'-04' %d.\h'+01'\c
.\}
.el \{\
.  sp -1
.  IP " %d." 4.2
.\}
`
	_roffTmplListOrderedItemEnd = `.RE
`
)

// List of roff template for unordered list.
const (
	_roffTmplListUnorderedItemBegin = `.RS 4
.ie n \{\
\h'-04'\(bu\h'+03'\c
.\}
.el \{\
.  sp -1
.  IP \(bu 2.3
.\}
`
	_roffTmplListUnorderedItemEnd = _roffTmplListOrderedItemEnd
)

// List of roff template for description list.
const (
	_roffTmplListDescriptionBegin = `%s
.RS 4
`
	_roffTmplListDescriptionEnd = `.RE
`
)

// roff macro for URL and MTO with link style.
const _roffTmplMacroLink = `.if \n[.g] \{\
.  mso www.tmac
.  am URL
.    ad l
.  .
.  am MTO
.    ad l
.  .
.  LINKSTYLE %s
.\}
`

const _roffTmplSection = `.%s %q
.sp
`
const _roffTmplSectionAuthor = `.SH "AUTHOR"
.sp
`

const _roffTmplUrl = `.URL "%s" "%s" "."
`
