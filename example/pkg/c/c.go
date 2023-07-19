package c

import (
	"fmt"
	"go-example/pkg/bridge"
)

type C struct {
	bridge.Bridge
}

func (c C) Print() {
	fmt.Println("C")
	c.PrintA()
}
func (c C) PrintA() {
	c.Bridge.Print()
}
