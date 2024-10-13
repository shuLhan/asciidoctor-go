package asciidoctor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseIncludeWithAbsolutePath(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)
	tdata, err = test.LoadData(`testdata/include_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var wd string

	wd, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	var fadoc = filepath.Join(wd, `testdata`, `include.adoc`)
	var doc *Document

	doc, err = Open(fadoc)
	if err != nil {
		t.Fatal(err)
	}

	var got bytes.Buffer

	doc.ToHTMLEmbedded(&got)

	var exp = string(tdata.Output[`include`])
	test.Assert(t, `ParseIncludeWithAbsolutePath`, exp, got.String())
}
