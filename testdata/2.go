package foo

import (
	"fmt"
)

type Foo struct {
	Bar string
	Boo int
	Arr []int
}

func (f *Foo) Bar() {
}

func (f Foo) Baz() error {
	return nil
}

func Qux() {
	f := Foo{}
	f.Arr = []int{0, 1, 2}
	x := 123
	fmt.Println("hello", f, x)
}
