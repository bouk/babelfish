package translate

import (
	"fmt"
	"mvdan.cc/sh/v3/syntax"
)

type UnsupportedError struct {
	Node syntax.Node
}

func (u *UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported: %#v", u.Node)
}

func unsupported(n syntax.Node) {
	panic(&UnsupportedError{n})
}
