// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package asciidoctor

import "github.com/shuLhan/share/lib/math/big"

const (
	Version = `0.3.1`
)

func init() {
	big.DefaultDigitPrecision = 6
}
