package asciidoctor

// List of passthrough substitutions.
const (
	passSubNone     int = 0
	passSubChar         = 1  // 'c'
	passSubQuote        = 2  // 'q'
	passSubAttr         = 4  // 'a'
	passSubRepl         = 8  // 'r'
	passSubMacro        = 16 // 'm'
	passSubPostRepl     = 32 // 'p'
	passSubNormal       = passSubChar | passSubQuote | passSubAttr | passSubRepl | passSubMacro | passSubPostRepl
	passSubVerbatim     = passSubChar
)
