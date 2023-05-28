# asciidoctor-go

Author: [Shulhan](mailto:ms@kilabit.info)

The asciidoctor-go is the Go module to parse the
[AsciiDoc markup](https://asciidoctor.org/docs/what-is-asciidoc)
and convert it into HTML5.

##  Documentation

[Go documentation](https://pkg.go.dev/git.sr.ht/~shulhan/asciidoctor-go).

### Specifications

During reverse engineering the AsciiDoc markup, we write the syntax, rules,
and format in
[SPECS](SPECS.html).

### Features

List of available formatting that are supported on current implementation.
Each supported feature is linked to official
[AsciiDoc Language Documentation](https://docs.asciidoctor.org/asciidoc/latest)
The numbered one is based on the old documentation.

* [Document header](https://docs.asciidoctor.org/asciidoc/latest/document/header/)
  * [Document title](https://docs.asciidoctor.org/asciidoc/latest/document/header/document/title/).
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
* Unordered Lists (See Notes below)
  * Nested
  * Complex List Content
  * Custom Markers
  * Checklist
* Ordered Lists
  * Nested
  * Numbering Styles
* Description List
  * Question and Answer Style List
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
  * ~~Nested tables~~
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
* Include Directive
  * Anatomy
  * Processing
  * File resolution
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

Supported metadata or attribute references,

* `author(_x)`
* `authorinitials(_x)`
* `doctitle`
* `email(_x)`
* `firstname(_x)`
* `idprefix`
* `idseparator`
* `lastname(_x)`
* `last-update-label`
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
* `table-caption`
* `title-separator`
* `version-label`


Additional metadata provides by this library,

* `author_names` - list of author full names separated by comma.

###  Notes

#### Unsupported markup

The following markup will not supported because its functionality is duplicate
or inconsistent with others markup, or not secure,

* Header
  * Subtitle partitioning.
    Rationale: duplicate with 14.1.2 the "Main: sub" format

* Tables
  * Nested tables.
    Rationale: nested table is not a good way to present information.
    Never should it be.
  * Using different cell separator

* Include Directive
  * Select Portions of a Document to Include.
    Rationale: the parser would need to know the language to be included and
    parse the whole source code to check for comments and tags.
  * Include Content from a URI.
    Rationale: security and unreliable network connections.
  * Caching URI Content


####  Unordered list item with hyphen

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

###  TODO

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
  * Partitioning large documents and using leveloffset
  * AsciiDoc vs non-AsciiDoc files
  * Normalize Block Indentation
  * Include a File Multiple Times in the Same Document
  * Using an Include in a List Item
* Text Substitutions
  * Macros
  * Incremental Substitutions
* Passthroughs
  * Passthrough Macros


####  Bugs

Unknown.
If you found one, please report it
[here](https://todo.sr.ht/~shulhan/asciidoctor-go).


####  Enhancements

Create tree that link Include directive.
Once the included files changes, the parent should be re-converted too.

```
Include Node
parent -> Parent Node.
```

###  Miscellaneous

[Changelog](CHANGELOG.html).

The following files compare the HTML generated by asciidoctor and this
library:

* [HTML file generated using asciidoctor](testdata/test.exp.html)
* [HTML file using this library](testdata/test.got.html)


## Development

[Repository](https://git.sr.ht/~shulhan/asciidoctor-go)::
Link to the source code.

[Mailing list](https://lists.sr.ht/~shulhan/asciidoctor-go)::
Link to discussion or where to send the patches.

[Issues](https://todo.sr.ht/~shulhan/asciidoctor-go)::
Link to open an issue or request for new feature.
