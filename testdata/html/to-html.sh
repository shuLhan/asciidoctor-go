#!/bin/sh

## Script to convert all adoc files to HTML without stylesheet.

asciidoctor -a stylesheet! -o header.exp.html header.adoc
asciidoctor -a stylesheet! -o preamble.exp.html preamble.adoc
