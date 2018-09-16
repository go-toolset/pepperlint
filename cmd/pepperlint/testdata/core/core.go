package main

import (
	"fmt"

	"github.com/go-toolset/pepperlint/cmd/pepperlint/testdata/deprecated"
)

// Foo using deprecated structure
type Foo deprecated.Deprecated

func deprecatedFunction(bar deprecated.Deprecated) deprecated.Deprecated {
	foo := deprecated.Deprecated{
		DeprecatedField: 1,
	}

	foo.DeprecatedField = 2
	v := moo(foo)
	if v == 0 {
		return foo // TODO
	}

	fmt.Println(foo)
	fmt.Println(foo.DeprecatedField)

	baz := &deprecated.Deprecated{} // TODO
	baz.DeprecatedOp()
	a := baz.DeprecatedOp()

	baz.DeprecatedPtrOp()
	b := baz.DeprecatedPtrOp()

	deprecated.DeprecatedFunction() // TODO

	return deprecated.Deprecated{}
}

func moo(foo deprecated.Deprecated) int32 {
	return foo.DeprecatedField
}
