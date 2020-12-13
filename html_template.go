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
<meta name="generator" content="asciidoctor-go">`
)

const (
	_htmlPreambleBegin = `
<div id="preamble">
<div class="sectionbody">`
)

//
// HTML templates for table of contents.
//
const (
	_htmlToCBegin = `
<div id="toc" class="%s">
<div id="toctitle">%s</div>`
)

//
// HTML templates for adminition block.
//
const (
	_htmlAdmonitionIconsFont = `
<i class="fa icon-%s" title=%q></i>`

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
<img src=%q alt=%q%s%s>
</div>`
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
)

//
// HTML templates for list description.
//
const (
	_htmlListDescriptionItemBegin = `
<dt class="hdlist1">%s</dt>
<dd>`

	_htmlListDescriptionItemQandABegin = `
<li>
<p><em>%s</em></p>`

	_htmlListDescriptionItemHorizontalBegin = `
<tr>
<td class="hdlist1">
%s
</td>
<td class="hdlist2">`
)
