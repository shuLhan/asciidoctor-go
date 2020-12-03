# asciidoctor-go

Shulhan <ms@kilabit.info>

The asciidoctor-go is the Go module to parse the
[AsciiDoc markup](https://asciidoctor.org/docs/what-is-asciidoc)
and convert into HTML5.

## Features

List of available formatting that are supported on current implementation,
numbered based on
[AsciiDoc user manual](https://asciidoctor.org/docs/user-manual/),

* 14. Header
  * 14.1. Document title
    * 14.1.1. doctitle attribute
    * 14.1.2. Document subtitle
    * 14.1.3. Document title visibility
  * 14.2. Author and email
  * 14.3. Revision number, date
  * 14.5. Metadata
    * 14.5.1. Description
    * 14.5.2. Keywords
* 15. Preamble
* 16. Sections
  * 16.1. Titles as HTML headings
  * 16.2. Auto-generated IDs
  * 16.3. Custom IDs
  * 16.4. Multiple Anchors
  * 16.5. Links
  * 16.6. Anchors
  * 16.7. Numbering
  * 16.8. Discrete headings
* 17. Blocks
  * 17.1. Title
  * 17.2. Metadata
  * 17.3. Delimited blocks
* 18. Paragraph
  * 18.1. Alignment
  * 18.2. Line breaks (" +\n")
* 19. Text formatting
  * 19.1. Bold and italic
  * 19.2. Quotation Marks and Apostrophes
  * 19.3. Subscript and Superscript
  * 19.4. Monospace
* 20. Unordered Lists
  * 20.1. Nested
  * 20.2. Complex List Content
* 21. Ordered Lists
  * 21.1. Nested
  * 21.2. Numbering Styles
* 22. Description List
  * 22.1. Question and Answer Style List
* 23. Tables
  * 23.1. Columns
  * 23.2. Column formatting
  * 23.3. Cell formatting
  * 23.4. Header row
  * 23.5. Footer row
  * 23.6. Table width
* 24. Horizontal Rules
  * 24.1. Markdown-style horizontal rules
* 25. Page Break
* 26. URLs
  * 26.1. Link to Relative Files
* 27. Cross References
  * 27.1. Automatic Anchors
  * 27.2. Defining an Anchor
  * 27.3. Internal Cross References
  * 27.5. Customizing the Cross Reference Text
* 29. Images
* 30. Video
  * 30.1. YouTube and Vimeo videos
  * 30.2. Supported Attributes
* 31. Audio
* 32. Admonition
* 33. Sidebar
* 34. Example
* 35. Prose Excerpts, Quotes and Verses
  * 35.1. Quote
  * 35.2. Verse
* 36. Comments
* 37. Text Substitutions
  * 37.1. Special Characters
  * 37.2. Quotes
  * 37.3. Attributes (reference)
  * 37.4. Replacements
  * 37.9. Preventing Substitutions
* 39. Listing Blocks
* 40. Passthroughs
  * 40.2. Passthrough Blocks
* 41. Open Blocks

Supported metadata or attribute references,

* `author(_x)`
* `authorinitials(_x)`
* `doctitle`
* `email(_x)`
* `firstname(_x)`
* `idprefix`
* `idseparator`
* `lastname(_x)`
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
* `title-separator`
* `version-label`


##  TODO

List of features which will be implemented,

* 16. Sections
  * 16.9. Section styles
* 18. Paragraph
  * 18.2. Line breaks
    * Per block "[%hardbreaks]"
    * All document ":hardbreaks:"
  * 18.3. Lead style
* 19. Text formatting
  * 19.5. Custom Styling With Attributes
* 20. Unordered Lists
  * 20.3. Custom Markers
  * 20.4. Checklist
* 22. Description List
  * Style on description label
* 23. Tables
* 27. Cross References
  * 27.6. Inter-document Cross References
* 28. Include Directive
* 37. Text Substitutions
  * 37.5. Macros
  * 37.8. Incremental Substitutions
* 40. Passthroughs
  * 40.1. Passthrough Macros


## Not supported

The following asciidoctor markup will not supported because its functionality
is duplicate with others markup,

* 14. Header
  * 14.4. Subtitle partitioning. Duplicate with 14.1.2 the "Main: sub" format.
