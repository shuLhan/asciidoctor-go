package asciidoctor

type parserInlineState struct {
	stack []int
}

func (pis *parserInlineState) push(c int) {
	pis.stack = append(pis.stack, c)
}

func (pis *parserInlineState) pop() (c int) {
	size := len(pis.stack)
	if size > 0 {
		c = pis.stack[size-1]
		pis.stack = pis.stack[:size-1]
	}
	return c
}

func (pis *parserInlineState) has(c int) bool {
	for _, r := range pis.stack {
		if r == c {
			return true
		}
	}
	return false
}
