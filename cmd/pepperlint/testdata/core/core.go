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
	v := moo(foo) // TODO
	if v == 0 {
		return foo
	}

	fmt.Println(foo)
	fmt.Println(foo.DeprecatedField)
	return deprecated.Deprecated{}
}

func moo(foo deprecated.Deprecated) int32 {
	return foo.DeprecatedField
}
