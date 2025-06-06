//go:generate go run main.go
package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"io"
	"os"
	"text/template"
)

//go:embed imports.go.tmpl
var tsImports string

//go:embed prepare_row.go.tmpl
var tsPrepareRow string

//go:embed process_field.go.tmpl
var tsProcessField string

//go:embed escape_chars.go.tmpl
var tsEscapeChars string

//go:embed write.go.tmpl
var tsWrite string

func parse(s string) *template.Template {
	t, err := template.New("").Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

func renderer[T any](w io.Writer) func(*template.Template, []T) {
	return func(t *template.Template, data []T) {
		if len(data) == 0 {
			if err := t.Execute(w, nil); err != nil {
				panic(err)
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				panic(err)
			}
			return
		}

		for _, d := range data {
			if err := t.Execute(w, d); err != nil {
				panic(err)
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	const dstFile = "../../../gen_strategies.go"

	defer func() {
		if r := recover(); r != nil {
			defer panic(r)
			os.Remove(dstFile)
			return
		}
	}()

	f, err := os.Create(dstFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			v := any(err)
			if r := recover(); r != nil {
				v = r
			}
			panic(v)
		}
	}()

	var buf bytes.Buffer

	_, err = buf.WriteString(`// Code generated by ./internal/cmd/generate/main.go DO NOT EDIT.` + "\n\n")
	if err != nil {
		panic(err)
	}

	// create imports section of source code
	{
		render := renderer[any](&buf)

		render(parse(tsImports), nil)
	}

	// render prepareRow strategies
	{
		t := parse(tsPrepareRow)

		type cfg struct {
			ClearMemoryAfterUse bool
		}

		render := renderer[cfg](&buf)

		render(t, []cfg{
			{ClearMemoryAfterUse: false},
			{ClearMemoryAfterUse: true},
		})
	}

	// render processField strategies
	{
		t := parse(tsProcessField)

		type cfg struct {
			EscapeSet           bool
			QuoteForced         bool
			ClearMemoryAfterUse bool
		}

		render := renderer[cfg](&buf)

		render(t, []cfg{
			{EscapeSet: false, QuoteForced: false, ClearMemoryAfterUse: false},
			{EscapeSet: true, QuoteForced: false, ClearMemoryAfterUse: false},
			{EscapeSet: false, QuoteForced: true, ClearMemoryAfterUse: false},
			{EscapeSet: true, QuoteForced: true, ClearMemoryAfterUse: false},
			{EscapeSet: false, QuoteForced: false, ClearMemoryAfterUse: true},
			{EscapeSet: true, QuoteForced: false, ClearMemoryAfterUse: true},
			{EscapeSet: false, QuoteForced: true, ClearMemoryAfterUse: true},
			{EscapeSet: true, QuoteForced: true, ClearMemoryAfterUse: true},
		})
	}

	// render EscapeChars strategies
	{
		t := parse(tsEscapeChars)

		type cfg struct {
			EscapeEnabled       bool
			ClearMemoryAfterUse bool
		}

		render := renderer[cfg](&buf)

		render(t, []cfg{
			{EscapeEnabled: false, ClearMemoryAfterUse: false},
			{EscapeEnabled: true, ClearMemoryAfterUse: false},
			{EscapeEnabled: false, ClearMemoryAfterUse: true},
			{EscapeEnabled: true, ClearMemoryAfterUse: true},
		})
	}

	// render writing strategies
	{
		t := parse(tsWrite)

		type cfg struct {
			ClearMemoryAfterUse bool
		}

		render := renderer[cfg](&buf)

		render(t, []cfg{
			{ClearMemoryAfterUse: false},
			{ClearMemoryAfterUse: true},
		})
	}

	// // for debugging
	// _, err = f.Write(buf.Bytes())
	// if err != nil {
	// 	panic(err)
	// } else {
	// 	return
	// }

	b, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}
