package a

import (
	"fmt"
	"go-example/pkg/b"
)

type A struct {
	b.B
}

func (a A) Print() {
	fmt.Println("A")
	a.PrintB()
}

func (a A) PrintB() {
	a.B.Print()
}
