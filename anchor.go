// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asciidoctor

//
// anchor contains label and counter for duplicate ID.
//
type anchor struct {
	label   string
	counter int
}
