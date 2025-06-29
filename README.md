# asciidoctor-go

The asciidoctor-go is the Go module to parse the
[AsciiDoc markup](https://asciidoctor.org/docs/what-is-asciidoc)
and convert it into HTML5.

For the front-end tooling that use this library to build static website
see
[ciigo](https://sr.ht/~shulhan/ciigo).


##  Documentation

See the
[Go module documentation](https://pkg.go.dev/git.sr.ht/~shulhan/asciidoctor-go)
for the API and examples on how to use this library to parse and render
Asciidoc file.

During reverse engineering the AsciiDoc markup, we write the syntax, rules,
and format in
[specification file](SPECS.html).


## Features

List of available formatting that are supported on current implementation.
Each supported feature is linked to official
[AsciiDoc Language Documentation](https://docs.asciidoctor.org/asciidoc/latest)

* [Document header](https://docs.asciidoctor.org/asciidoc/latest/document/header/)
  * [Document title](https://docs.asciidoctor.org/asciidoc/latest/document/title/).
    This including meta `doctitle`, `showtitle!` and subtitle.
  * [Author information](https://docs.asciidoctor.org/asciidoc/latest/document/author-information/)
  * [Revision information](https://docs.asciidoctor.org/asciidoc/latest/document/revision-information/)
  * [Metadata](https://docs.asciidoctor.org/asciidoc/latest/document/metadata/)
* [Preamble](https://docs.asciidoctor.org/asciidoc/latest/blocks/preamble-and-lead/)
* Sections
  * Titles as HTML headings
  * Auto-generated IDs
  * Custom IDs
  * Multiple Anchors
  * Links
  * Anchors
  * Numbering
  * Discrete headings
* Blocks
  * Title
  * Metadata
  * Delimited blocks
* Paragraph
  * Alignment
  * Line breaks (" +\n")
  * Lead style
* Text formatting
  * Bold and italic
  * Quotation Marks and Apostrophes
  * Subscript and Superscript
  * Monospace
* Lists &#x221A;
  * [Unordered Lists](https://docs.asciidoctor.org/asciidoc/latest/lists/unordered/) &#x221A; (see Notes below)
    * Basic unordered list &#x221A;
    * Nested unordered list &#x221A;
    * Markers &#x221A;
      * Default markers &#x221A;
      * Custom markers &#x221A;
  * [Ordered Lists](https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/) &#x221A;
    * Basic ordered list &#x221A;
    * Nested ordered list &#x221A;
    * Number styles &#x221A;
  * [Checklists](https://docs.asciidoctor.org/asciidoc/latest/lists/checklist/) &#x221A;
  * [Separating Lists](https://docs.asciidoctor.org/asciidoc/latest/lists/separating/) &#x221A;
* Description Lists
  * [Description Lists](https://docs.asciidoctor.org/asciidoc/latest/lists/description/)
    * Basic description list &#x221A;
    * Mixing lists &#x221A;
    * Nested description list &#x221A;
  * Question and Answer Lists &#x221A;
* Tables
  * Columns
  * Column formatting
  * Cell formatting
  * Header row
  * Footer row
  * Table width
  * Table borders
  * Striping
  * Orientation
  * Table caption
  * Escaping the Cell Separator
* Horizontal Rules
  * Markdown-style horizontal rules
* Page Break
* URLs
  * Link to Relative Files
* Cross References
  * Automatic Anchors
  * Defining an Anchor
  * Internal Cross References
  * Customizing the Cross Reference Text
* [Footnotes](https://docs.asciidoctor.org/asciidoc/latest/macros/footnote/)
* [Includes](https://docs.asciidoctor.org/asciidoc/latest/directives/include/)
* Images
* Video
  * YouTube and Vimeo videos
  * Supported Attributes
* Audio
* Admonition
* Sidebar
* Example
* Prose Excerpts, Quotes and Verses
  * Quote
  * Verse
* Comments
* Text Substitutions
  * Special Characters
  * Quotes
  * Attributes (reference)
  * Replacements
  * Preventing Substitutions
* Listing Blocks
* Passthroughs
  * Passthrough Blocks
* Open Blocks
* Predefined Attributes for Character Replacements

Supported document attribute references,

* `author(_x)`
* `authorinitials(_x)`
* `docdir`
* `doctitle`
* `email(_x)`
* `firstname(_x)`
* `idprefix`
* `idseparator`
* `lastname(_x)`
* `last-update-label`
* [`leveloffset`](https://docs.asciidoctor.org/asciidoc/latest/directives/include-with-leveloffset/).
Only on document attributes, not on include directive.
* `middlename(_x)`
* `nofooter`
* `noheader`
* `revdate`
* `revnumber`
* `revremark`
* `sectids`
* `sectnumlevels`
* `sectnums`
* `showtitle`
* `stylesheet`
* `table-caption`
* `title-separator`
* `version-label`


Additional document attribute provides by this library,

* `author_names` - list of author full names separated by comma.


##  Notes

### Unsupported markup

The following markup will not supported either because its functionality is
duplicate, or inconsistent with others markup, or not secure,

* Header
  * Subtitle partitioning.
    Rationale: duplicate with the "Main: sub" format

* Tables
  * Nested tables.
    Rationale: nested table is not a good way to present information,
    never should it be.
  * Using different cell separator

* Includes
  * [Include Content by Tagged Regions](https://docs.asciidoctor.org/asciidoc/latest/directives/include-tagged-regions/)
    Rationale: the parser would need to know the language to be included and
    parse the whole source code to check for comments and tags.
  * [Include Content by URI](https://docs.asciidoctor.org/asciidoc/latest/directives/include-uri/)
    Rationale: security and unreliable network connections.


###  Unordered list item with hyphen

The unordered list item with hyphen ('-') cause too much confusion and
inconsistency.

Case one, given the following markup,

```
- Item 1
+
"A line
of quote"
-- Author
```

Is the "Author" the sub item in list or we are parsing author of quote
paragraph?

Case two, the writer want to write em dash (`&#8212;` in HTML Unicode) but
somehow the editor wrap it and start in new line.

As a reminder, the official documentation only recommend using hyphen for
simple list item

> You should reserve the hyphen for lists that only have a single level
> because the hyphen marker (-) doesn’t work for nested lists.
> -- <https://docs.asciidoctor.org/asciidoc/latest/lists/unordered/#basic-unordered-list>


##  TODO

List of features which may be implemented,

* Sections
  * Section styles
* Paragraph
  * Line breaks
    * Per block "[%hardbreaks]"
    * All document ":hardbreaks:"
* Text formatting
  * Custom Styling With Attributes
* Tables
  * Delimiter-Separated Values
* Cross References
  * Inter-document Cross References
* Include Directive
  * Offset Section Levels
  * Indent Included Content
  * Use an Include File Multiple Times
  * Include List Item Content
  * Include Content by Line Ranges
* Text Substitutions
  * Macros
  * Incremental Substitutions
* Passthroughs
  * Passthrough Macros


Future enhancements,

(1) Create tree that link Include directive.
Once the included files changes, the parent should be re-converted too.


##  Bugs

Unknown.
If you found one, please report it
[here](https://todo.sr.ht/~shulhan/asciidoctor-go).


## Development

[Changelog](https://kilabit.info/project/asciidoctor-go/CHANGELOG.html)::
List of each releases and their changes.

[Repository](https://git.sr.ht/~shulhan/asciidoctor-go)::
Link to the source code.

[Mailing list](https://lists.sr.ht/~shulhan/asciidoctor-go)::
Link to discussion or where to send the patches.

[Issues](https://todo.sr.ht/~shulhan/asciidoctor-go)::
Link to open an issue or request for new feature.


## License

Copyright (C) 2021 Shulhan [&lt;ms@kilabit.info&gt;](mailto:ms@kilabit.info)

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU General Public License as published by the Free
Software Foundation, either version 3 of the License, or (at your option)
any later version.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <http://www.gnu.org/licenses/>.
