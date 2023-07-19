package b

import (
	"fmt"
	"go-example/pkg/c"
)

type B struct {
	c.C
}

func (b B) Print() {
	fmt.Println("B")
	b.PrintC()
}
func (b B) PrintC() {
	b.C.Print()
}
