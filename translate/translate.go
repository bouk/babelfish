package translate

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"mvdan.cc/sh/syntax"
)

type Translator struct {
	buf bytes.Buffer
}

func NewTranslator() *Translator {
	return &Translator{}
}

func (t *Translator) WriteTo(w io.Writer) (int64, error) {
	return t.buf.WriteTo(w)
}

func (t *Translator) File(f *syntax.File) error {
	for _, stmt := range f.Stmts {
		if err := t.Stmt(stmt); err != nil {
			return err
		}
		t.str("\n")
	}

	for _, comment := range f.Last {
		if err := t.Comment(&comment); err != nil {
			return err
		}
	}

	return nil
}

func (t *Translator) Stmt(s *syntax.Stmt) error {
	for _, comment := range s.Comments {
		if err := t.Comment(&comment); err != nil {
			return err
		}
	}

	return t.Command(s.Cmd)
}

type UnsupportedError struct {
	Node syntax.Node
}

func (u *UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported: %#v", u.Node)
}

func (t *Translator) Command(c syntax.Command) error {
	switch c := c.(type) {
	case *syntax.ArithmCmd:
		return &UnsupportedError{c}
	case *syntax.BinaryCmd:
		return &UnsupportedError{c}
	case *syntax.Block:
		return &UnsupportedError{c}
	case *syntax.CallExpr:
		return t.CallExpr(c)
	case *syntax.CaseClause:
		return &UnsupportedError{c}
	case *syntax.CoprocClause:
		return &UnsupportedError{c}
	case *syntax.DeclClause:
		return t.DeclClause(c)
	case *syntax.ForClause:
		return &UnsupportedError{c}
	case *syntax.FuncDecl:
		return &UnsupportedError{c}
	case *syntax.IfClause:
		return &UnsupportedError{c}
	case *syntax.LetClause:
		return &UnsupportedError{c}
	case *syntax.Subshell:
		return &UnsupportedError{c}
	case *syntax.TestClause:
		return &UnsupportedError{c}
	case *syntax.TimeClause:
		return &UnsupportedError{c}
	case *syntax.WhileClause:
		return &UnsupportedError{c}
	default:
		return &UnsupportedError{c}
	}
}

func (t *Translator) CallExpr(c *syntax.CallExpr) error {
	if len(c.Args) == 0 {
		// assignment
		for n, a := range c.Assigns {
			if n > 0 {
				t.str("; ")
			}

			t.printf("set %s ", a.Name.Value)
			if err := t.Word(a.Value); err != nil {
				return err
			}
		}

		return nil
	} else {
		// call
		if len(c.Assigns) > 0 {
			t.str("env ")
			for _, a := range c.Assigns {
				if a.Value == nil {
					t.printf("-u %s ", a.Name.Value)
				} else {
					t.printf("%s=", a.Name.Value)
					if err := t.Word(a.Value); err != nil {
						return err
					}
					t.str(" ")
				}
			}
		}

		for i, a := range c.Args {
			if i > 0 {
				t.str(" ")
			}
			if err := t.Word(a); err != nil {
				return err
			}
		}
		return nil
	}
}

func (t *Translator) DeclClause(c *syntax.DeclClause) error {
	var prefix string
	if c.Variant != nil {
		switch c.Variant.Value {
		case "export":
			prefix = " -gx"
		case "local":
			prefix = " -l"
		default:
			return &UnsupportedError{c}
		}
	}

	for _, a := range c.Assigns {
		t.printf("set%s %s ", prefix, a.Name.Value)
		if a.Value == nil {
			t.printf("$%s", a.Name.Value)
		} else {
			if err := t.Word(a.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Translator) Word(w *syntax.Word) error {
	for _, part := range w.Parts {
		if err := t.WordPart(part); err != nil {
			return err
		}
	}
	return nil
}

func (t *Translator) WordPart(wp syntax.WordPart) error {
	switch wp := wp.(type) {
	case *syntax.Lit:
		t.str(wp.Value)
		return nil
	case *syntax.SglQuoted:
		return t.escapedString(wp.Value)
	case *syntax.DblQuoted:
		for _, part := range wp.Parts {
			switch part := part.(type) {
			case *syntax.Lit:
				if err := t.escapedString(part.Value); err != nil {
					return err
				}
			default:
				if err := t.WordPart(part); err != nil {
					return err
				}
			}
		}
		return nil
	case *syntax.ParamExp:
		if wp.Short {
			t.printf("$%s", wp.Param.Value)
			return nil
		}
		if !wp.Excl && !wp.Length && !wp.Width {
			t.printf("{$%s}", wp.Param.Value)
			return nil
		}
		return &UnsupportedError{wp}
	case *syntax.CmdSubst:
		t.str("(")
		for i, s := range wp.Stmts {
			if i > 0 {
				t.str("; ")
			}

			if err := t.Stmt(s); err != nil {
				return err
			}
		}
		t.str(")")
		return nil
	case *syntax.ArithmExp:
		return &UnsupportedError{wp}
	case *syntax.ProcSubst:
		return &UnsupportedError{wp}
	case *syntax.ExtGlob:
		return &UnsupportedError{wp}
	default:
		return &UnsupportedError{wp}
	}
}

var stringReplacer = strings.NewReplacer("\\", "\\\\", "'", "\\'")

func (t *Translator) escapedString(literal string) error {
	t.str("'")
	stringReplacer.WriteString(&t.buf, literal)
	t.str("'")
	return nil
}

func (t *Translator) Comment(c *syntax.Comment) error {
	t.printf("#%s\n", c.Text)
	return nil
}

func (t *Translator) str(s string) {
	t.buf.WriteString(s)
}

func (t *Translator) printf(format string, arg ...interface{}) {
	fmt.Fprintf(&t.buf, format, arg...)
}
