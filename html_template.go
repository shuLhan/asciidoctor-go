// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

import (
	"strings"
	"text/template"
)

func (doc *Document) createHTMLTemplate() (tmpl *template.Template, err error) {
	imageCounter := 0
	exampleCounter := 0

	tmpl, err = template.New("HTML").Funcs(map[string]interface{}{
		// docAttribute access the global document attributes using
		// specific key.
		"docAttribute": func(key string) string {
			return doc.attributes[key]
		},
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
	}).Parse(`
{{- define "BEGIN" -}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="generator" content="ciigo">
	{{- if .Author}}
<meta name="author" content="{{.Author}}">
	{{- end -}}
	{{- if .Title}}
<title>{{.Title}}</title>
	{{- end}}
<style>

</style>
</head>
<body class="article">
{{- end -}}

{{- define "END"}}
</div>
<div id="footer">
<div id="footer-text">
	{{- if .RevNumber}}
Version {{.RevNumber}}<br>
	{{- end}}
Last updated {{.LastUpdated}}
</div>
</div>
</body>
</html>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_HEADER"}}
<div id="header">
	{{- if .Title}}
<h1>{{.Title}}</h1>
	{{- end}}
<div class="details">
	{{- if .Author}}
<span id="author" class="author">{{.Author}}</span><br>
	{{- end}}
	{{- if .RevNumber}}
<span id="revnumber">version {{.RevNumber}}{{.RevSeparator}}</span>
	{{- end}}
	{{- if .RevDate}}
<span id="revdate">{{.RevDate}}</span>
	{{- end}}
</div>
</div>
<div id="content">
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_PREAMBLE"}}
<div id="preamble">
<div class="sectionbody">
{{- end}}

{{- define "END_PREAMBLE"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SECTION_L1"}}
<div class="sect1">
<h2 id="{{.GenerateID .Content}}">{{- .Content -}}</h2>
<div class="sectionbody">
{{- end}}
{{- define "END_SECTION_L1"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SECTION_L2"}}
<div class="sect2">
<h3 id="{{.GenerateID .Content}}">{{- .Content -}}</h3>
{{- end}}
{{- define "END_SECTION"}}
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SECTION_L3"}}
<div class="sect3">
<h4 id="{{.GenerateID .Content}}">{{- .Content -}}</h4>
<div class="sectionbody">
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SECTION_L4"}}
<div class="sect4">
<h5 id="{{.GenerateID .Content}}">{{- .Content -}}</h5>
<div class="sectionbody">
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SECTION_L5"}}
<div class="sect5">
<h6 id="{{.GenerateID .Content}}">{{- .Content -}}</h6>
<div class="sectionbody">
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_TITLE"}}
	{{- with $title := .Title}}
<div class="title">{{$title}}</div>
	{{- end}}
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_PARAGRAPH"}}
<div class="paragraph {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<p>
{{- end}}
{{- define "END_PARAGRAPH"}}</p>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_LITERAL"}}
<div class="literalblock {{- .Classes}}">
<div class="content">
<pre>{{.Content -}}</pre>
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_LISTING"}}
<div class="listingblock {{- .Classes}}">
<div class="content">
<pre>{{.Content -}}</pre>
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_LIST_ORDERED"}}
{{- $class := .GetListOrderedClass}}
{{- $type := .GetListOrderedType}}
<div class="olist {{$class}} {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<ol class="{{$class}}"{{- if $type}} type="{{$type}}"{{end}}>
{{- end}}
{{- define "END_LIST_ORDERED"}}
</ol>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_LIST_UNORDERED"}}
<div class="ulist {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<ul>
{{- end}}
{{define "END_LIST_UNORDERED"}}
</ul>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_LIST_DESCRIPTION"}}
	{{- if .IsStyleQandA}}
<div class="qlist qanda {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<ol>
	{{- else if .IsStyleHorizontal}}
<div class="hdlist {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<table>
	{{- else}}
<div class="dlist {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<dl>
	{{- end}}
{{- end}}
{{- define "END_LIST_DESCRIPTION"}}
	{{- if .IsStyleQandA}}
</ol>
	{{- else if .IsStyleHorizontal}}
</table>
	{{- else}}
</dl>
	{{- end}}
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_LIST_ITEM"}}
<li>
<p>{{- .Content -}}</p>
{{- end}}
{{- define "END_LIST_ITEM"}}
</li>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_LIST_DESCRIPTION_ITEM"}}
	{{- if .IsStyleQandA}}
<li>
<p><em>{{.Label}}</em></p>
	{{- else if .IsStyleHorizontal}}
<tr>
<td class="hdlist1">
{{.Label}}
</td>
<td class="hdlist2">
	{{- else}}
<dt class="hdlist1">{{- .Label -}}</dt>
<dd>
	{{- end}}
	{{- with $content := .Content}}
<p>{{- $content -}}</p>
	{{- end}}
{{- end}}
{{- define "END_LIST_DESCRIPTION_ITEM"}}
	{{- if .IsStyleQandA}}
</li>
	{{- else if .IsStyleHorizontal}}
</td>
</tr>
	{{- else}}
</dd>
	{{- end}}
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "HORIZONTAL_RULE"}}
<hr>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "PAGE_BREAK"}}
<div style="page-break-after: always;"></div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_IMAGE"}}
<div class="imageblock {{- .Classes}}">
<div class="content">
<img src="{{.Content}}" alt="{{.Attrs.alt}}"
	{{- with $w := .Attrs.width}} width="{{$w}}"{{end}}
	{{- with $h := .Attrs.height}} height="{{$h}}"{{end}}>
</div>
{{- with $caption := .Title}}
<div class="title">Figure {{imageCounter}}. {{$caption}}</div>
{{- end}}
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_BLOCK_OPEN"}}
<div class="openblock {{- .Classes}}">
{{- template "BLOCK_TITLE" .}}
<div class="content">
{{- end}}

{{- define "END_BLOCK_OPEN"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BLOCK_VIDEO"}}
<div class="videoblock">
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
<div class="audioblock">
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
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_ADMONITION"}}
<div class="admonitionblock {{- .Classes}}">
<table>
<tr>
<td class="icon">
	{{- if eq (docAttribute "icons") "font"}}
<i class="fa icon-{{toLower .Classes}}" title="{{.Label}}"></i>
	{{- else}}
<div class="title">{{.Label}}</div>
	{{- end}}
</td>
<td class="content">
{{ with $content := .Content }}
{{$content}}
	{{- end}}
{{- end}}
{{- define "END_ADMONITION"}}
</td>
</tr>
</table>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_SIDEBAR"}}
<div class="sidebarblock {{- .Classes}}">
<div class="content">
{{- template "BLOCK_TITLE" .}}
{{- end}}
{{- define "END_SIDEBAR"}}
</div>
</div>
{{- end}}
{{/*----------------------------------------------------------------------*/}}
{{- define "BEGIN_EXAMPLE"}}
<div class="exampleblock {{- .Classes}}">
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
<div class="quoteblock {{- .Classes}}">
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
<div class="verseblock {{- .Classes}}">
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
{{/*----------------------------------------------------------------------*/}}

`)
	return tmpl, err
}
