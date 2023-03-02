// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

// Package asciidoctor-go is the Go module to parse the [AsciiDoc markup].
// Its currently support converting the asciidoc to HTML5.
//
// [AsciiDoc markup]: https://asciidoctor.org/docs/what-is-asciidoc
package asciidoctor

import "github.com/shuLhan/share/lib/math/big"

const (
	// Version of this module.
	Version = `0.4.1`

	_lf = "\n"
)

func init() {
	big.DefaultDigitPrecision = 6
}
