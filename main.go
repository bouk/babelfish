package main // import "bou.ke/babelfish"

import (
	"bou.ke/babelfish/translate"
	"flag"
	"fmt"
	"io"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"path/filepath"
)

type Options struct {
	Dump bool
}

func perform(name string, in io.Reader) error {
	out := os.Stdout
	errOut := os.Stderr
	p := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
	output, err := p.Parse(in, name)
	if err != nil {
		return err
	}

	t := translate.NewTranslator()

	loc := os.Args[0]
	// If the file path is relative, make it absolute
	if len(loc) > 0 && loc[0] == '.' {
		if wd, err := os.Getwd(); err == nil {
			loc = filepath.Join(wd, loc)
		}
	}
	t.BabelfishLocation(loc)

	err = t.File(output)
	if err, _ := err.(*translate.UnsupportedError); err != nil {
		syntax.NewPrinter().Print(errOut, err.Node)
		fmt.Fprintln(errOut)
		syntax.DebugPrint(errOut, err.Node)
		fmt.Fprintln(errOut)
	}
	if err == nil {
		_, err = t.WriteTo(out)
	}
	return err
}

func do() error {
	var o Options
	flag.BoolVar(&o.Dump, "dump", false, "Dump the AST")
	flag.Parse()

	f := os.Stdin
	if o.Dump {
		p := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
		output, err := p.Parse(f, f.Name())
		if err != nil {
			return err
		}
		syntax.DebugPrint(os.Stderr, output)
		fmt.Fprintln(os.Stderr)
		return nil
	}
	return perform(f.Name(), f)
}

func main() {
	if err := do(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
