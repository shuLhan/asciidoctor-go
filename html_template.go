// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

//
// HTML templates for head, meta attributes, and footers.
//
const (
	_htmlBegin = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="generator" content="ciigo">`

	_htmlMetaAuthor = `
<meta name="author" content="%s">`

	_htmlMetaDescription = `
<meta name="description" content="%s">`

	_htmlMetaKeywords = `
<meta name="keywords" content="%s">`

	_htmlHeadTitle = `
<title>%s</title>`

	_htmlHeadStyle = `
<style>

</style>`

	_htmlBodyBegin = `
</head>
<body class="%s">`

	_htmlFooterBegin = `
<div id="footer">
<div id="footer-text">`

	_htmlFooterVersion = `
Version %s<br>`

	_htmlFooterLastUpdated = `
Last updated %s`

	_htmlFooterEnd = `
</div>
</div>`

	_htmlBodyEnd = `
</body>
</html>`
)

//
// HTML templates for document header.
//
const (
	_htmlHeaderBegin = `
<div id="header">`

	_htmlHeaderTitleBegin = `
<h1>`
	_htmlHeaderTitleEnd = `</h1>`

	_htmlHeaderDetail = `
<div class="details">`
	_htmlHeaderDetailAuthor = `
<span id="author" class="author">%s</span><br>`
	_htmlHeaderDetailRevNumber = `
<span id="revnumber">version %s%s</span>`
	_htmlHeaderDetailRevDate = `
<span id="revdate">%s</span>`
	_htmlHeaderDetailEnd = `
</div>`

	_htmlHeaderEnd = `
</div>`
)

//
// HTML templates for content.
//
const (
	_htmlContentBegin = `
<div id="content">`

	_htmlContentEnd = `
</div>`
)

const (
	_htmlPreambleBegin = `
<div id="preamble">
<div class="sectionbody">`

	_htmlSection = `
<div class="%s">
<%s id="%s">`
)

//
// HTML templates for table of contents.
//
const (
	_htmlToCBegin = `
<div id="toc" class="%s">
<div id="toctitle">%s</div>`

	_htmlToCEnd = `
</div>`
)

//
// HTML templates for adminition block.
//
const (
	_htmlAdmonitionIconsFont = `
<i class="fa icon-%s" title="%s"></i>`

	_htmlAdmonitionContent = `
</td>
<td class="content">
%s`

	_htmlAdmonitionEnd = `
</td>
</tr>
</table>
</div>`
)

const (
	// Parameters: src, autoplay, controls, loop.
	_htmlBlockAudio = `
<audio src="%s"%s%s%s>
Your browser does not support the audio tag.
</audio>
</div>
</div>`
)

//
// HTML templates for block image.
//
const (
	_htmlBlockImage = `
<div class="content">
<img src="%s" alt="%s"%s%s>
</div>`

	_htmlBlockImageTitle = `
<div class="title">Figure %d. %s</div>`

	_htmlBlockImageEnd = `
</div>`
)

const (
	_htmlBlockQuoteBegin = `
<blockquote>
%s`

	_htmlBlockQuoteEnd = `
</blockquote>`

	_htmlBlockVerse = `
<pre class="content">%s`

	_htmlBlockVerseEnd = `</pre>`

	_htmlQuoteAuthor = `
<div class="attribution">
&#8212; %s`

	_htmlQuoteCitation = `<br>
<cite>%s</cite>`
)

// Block video.
const (
	// List of parameters in order: src, width, height, poster, controls,
	// autoplay, loop.
	_htmlBlockVideo = `
<video src="%s"%s%s%s%s%s%s>
Your browser does not support the video tag.
</video>`

	// List of parameters in order: width, height, src, allowfullscreen.
	_htmlBlockVideoYoutube = `
<iframe%s%s src="%s" frameborder="0"%s></iframe>`

	// List of parameters in order: width, height, src.
	_htmlBlockVideoVimeo = `
<iframe%s%s src="%s" frameborder="0"></iframe>`
)

const (
	_htmlBlockLiteralContent = `
<div class="content">
<pre>%s</pre>
</div>
</div>`

	_htmlBlockContent = `
<div class="content">`

	_htmlBlockExampleTitle = `
<div class="title">Example %d. %s</div>`

	_htmlBlockTitle = `
<div class="title">%s</div>`

	_htmlBlockEnd = `
</div>
</div>`
)

//
// Inline image.
//
const (
	_htmlInlineImage      = `<span class="%s">`
	_htmlInlineImageLink  = `<a class="image" href="%s">`
	_htmlInlineImageImage = `<img src="%s" alt="%s"%s%s>`
)

//
// HTML templates for list description.
//
const (
	_htmlListDescriptionItemBegin = `
<dt class="hdlist1">%s</dt>
<dd>`
	_htmlListDescriptionItemEnd = `
</dd>`

	_htmlListDescriptionItemQandABegin = `
<li>
<p><em>%s</em></p>`
	_htmlListDescriptionItemQandAEnd = `
</li>`

	_htmlListDescriptionItemHorizontalBegin = `
<tr>
<td class="hdlist1">
%s
</td>
<td class="hdlist2">`
	_htmlListDescriptionItemHorizontalEnd = `
</td>
</tr>`
)

//
// HTML templates for inline markup.
//
const (
	_htmlCrossReference   = `<a href="#%s">%s</a>`
	_htmlHorizontalRule   = "\n<hr>"
	_htmlInlineID         = "<a id=\"%s\"></a>"
	_htmlInlineIDShort    = `<span id="%s">%s`
	_htmlInlineIDShortEnd = `</span>`
	_htmlPageBreak        = `
<div style="page-break-after: always;"></div>`
)
