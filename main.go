package main

import (
	"bou.ke/babelfish/translate"
	"flag"
	"fmt"
	"mvdan.cc/sh/syntax"
	"os"
)

type Options struct {
	Dump bool
}

func do() error {
	var o Options
	flag.BoolVar(&o.Dump, "dump", false, "Dump the AST")
	flag.Parse()
	f := os.Stdin
	p := syntax.NewParser(syntax.KeepComments, syntax.Variant(syntax.LangBash))
	output, err := p.Parse(f, f.Name())
	if err != nil {
		return err
	}

	if o.Dump {
		syntax.DebugPrint(os.Stderr, output)
		fmt.Fprintln(os.Stderr)
		return nil
	}

	t := translate.NewTranslator()
	err = t.File(output)
	if err, _ := err.(*translate.UnsupportedError); err != nil {
		fmt.Fprintln(os.Stderr)
		syntax.DebugPrint(os.Stderr, err.Node)
		fmt.Fprintln(os.Stderr)
	}
	if err == nil {
		_, err = t.WriteTo(os.Stdout)
	}
	return err
}

func main() {
	if err := do(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
