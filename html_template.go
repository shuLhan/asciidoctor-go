// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"strings"
	"text/template"
)

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
// HTML templates for table of contens.
//
const (
	_htmlToCBegin = `
<div id="toc" class="%s">
<div id="toctitle">%s</div>`

	_htmlToCEnd = `
</div>`
)

const (
	_htmlAdmonitionIconsFont = `
<i class="fa icon-%s" title="%s"></i>`

	_htmlAdmonitionTitle = `
<div class="title">%s</div>`

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
	_htmlBlockLiteralContent = `
<div class="content">
<pre>%s</pre>
</div>
</div>`

	_htmlBlockTitle = `
<div class="title">%s</div>`
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

func (doc *Document) createHTMLTemplate() (tmpl *template.Template, err error) {
	imageCounter := 0
	exampleCounter := 0

	tmpl, err = template.New("HTML").Funcs(map[string]interface{}{
		"exampleCounter": func() int {
			exampleCounter++
			return exampleCounter
		},
		"imageCounter": func() int {
			imageCounter++
			return imageCounter
		},
		"toLower": func(s string) string {
			return strings.ToLower(strings.TrimSpace(s))
		},
		"trimSpace": func(s string) string {
			return strings.TrimSpace(s)
		},
	}).Parse(`
{{- define "BLOCK_TITLE"}}
	{{- with $title := .Title}}
<div class="title">{{$title}}</div>
	{{- end}}
{{- end}}


{{- define "BLOCK_IMAGE"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "imageblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
<div class="content">
<img src="{{.Attrs.src}}" alt="{{.Attrs.alt}}"
	{{- with $w := .Attrs.width}} width="{{$w}}"{{end}}
	{{- with $h := .Attrs.height}} height="{{$h}}"{{end}}>
</div>
{{- with $caption := .Title}}
<div class="title">Figure {{imageCounter}}. {{$caption}}</div>
{{- end}}
</div>
{{- end}}


{{- define "INLINE_IMAGE" -}}
{{- $link := .Attrs.link -}}
<span
	{{- with $c := printf "image %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- with $link}}<a class="image" href="{{$link}}">{{end -}}
<img src="{{.Attrs.src}}" alt="{{.Attrs.alt}}"
	{{- with $w := .Attrs.width}} width="{{$w}}"{{end}}
	{{- with $h := .Attrs.height}} height="{{$h}}"{{end}}>
{{- with $link}}</a>{{end -}}
</span>
{{- end}}
{{/*----------------------------------------------------------------------*/}}

{{- define "BEGIN_BLOCK_OPEN"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "openblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- template "BLOCK_TITLE" .}}
<div class="content">
{{- end}}

{{- define "END_BLOCK_OPEN"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_VIDEO"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}} class="videoblock">
{{- template "BLOCK_TITLE" .}}
<div class="content">
	{{- if .Attrs.youtube}}
<iframe
		{{- with $w := .Attrs.width}} width="{{$w}}" {{- end}}
		{{- with $h := .Attrs.height}} height="{{$h}}" {{- end}} src="{{.GetVideoSource}}" frameborder="0"
		{{- if not .Attrs.nofullscreen}} allowfullscreen{{end}}></iframe>
	{{- else if .Attrs.vimeo}}
<iframe
		{{- with $w := .Attrs.width}} width="{{$w}}" {{- end}}
		{{- with $h := .Attrs.height}} height="{{$h}}" {{- end }} src="{{.GetVideoSource}}" frameborder="0"></iframe>
	{{- else}}
<video src="{{.GetVideoSource}}"
		{{- with $w := .Attrs.width}} width="{{$w}}" {{- end}}
		{{- with $h := .Attrs.height}} height="{{$h}}" {{- end -}}
		{{- if .Attrs.poster}} poster="{{.Attrs.poster}}"{{end -}}
		{{- if not .Attrs.nocontrols}} controls{{end -}}
		{{- if .Attrs.autoplay}} autoplay{{end -}}
		{{- if .Attrs.loop}} loop{{end -}}
>
Your browser does not support the video tag.
</video>
	{{- end}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_AUDIO"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "audioblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- template "BLOCK_TITLE" .}}
<div class="content">
<audio src="{{.Attrs.src}}"
	{{- if .Opts.autoplay}} autoplay{{end}}
	{{- if eq .Opts.controls "1"}} controls{{end}}
	{{- if .Opts.loop}} loop{{end}}>
Your browser does not support the audio tag.
</audio>
</div>
</div>
{{- end}}


{{- define "BEGIN_SIDEBAR"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "sidebarblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
<div class="content">
{{- template "BLOCK_TITLE" .}}
{{- end}}
{{- define "END_SIDEBAR"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_EXAMPLE"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "exampleblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- with $caption := .Title}}
<div class="title">Example {{exampleCounter}}. {{$caption}}</div>
{{- end}}
<div class="content">
{{- end}}
{{- define "END_EXAMPLE"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_QUOTE"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "quoteblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- with $caption := .Title}}
<div class="title">{{$caption}}</div>
{{- end}}
<blockquote>
{{.Content}}
{{- end}}

{{- define "END_QUOTE"}}
{{- $author := .QuoteAuthor}}
{{- $citation := .QuoteCitation}}
</blockquote>
	{{- if $author}}
<div class="attribution">
&#8212; {{$author}}{{if $citation}}<br>{{end}}
	{{- end}}
	{{- if $citation}}
<cite>{{$citation}}</cite>
	{{- end}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_VERSE"}}
<div
	{{- if .ID}} id="{{.ID}}"{{end}}
	{{- with $c := printf "verseblock %s" .Classes | trimSpace}} class="{{$c}}"{{end -}}
>
{{- with $caption := .Title}}
<div class="title">{{$caption}}</div>
{{- end}}
<pre class="content">{{.Content}}
{{- end}}
{{- define "END_VERSE"}}
{{- $author := .QuoteAuthor}}
{{- $citation := .QuoteCitation -}}
</pre>
	{{- if $author}}
<div class="attribution">
&#8212; {{$author}}{{if $citation}}<br>{{end}}
	{{- end}}
	{{- if $citation}}
<cite>{{$citation}}</cite>
	{{- end}}
</div>
</div>
{{- end}}
`)
	return tmpl, err
}
