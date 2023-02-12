#!/bin/sh

## Script to convert all adoc files to HTML without stylesheet.

asciidoctor -a stylesheet! -o header.exp.html header.adoc
asciidoctor -a stylesheet! -o preamble.exp.html preamble.adoc
asciidoctor -a stylesheet! -o substitutions.exp.html substitutions.adoc
asciidoctor -a stylesheet! -o text_formatting.exp.html text_formatting.adoc
