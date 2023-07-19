package main

import (
	"go-example/pkg/a"
	"go-example/pkg/b"
	"go-example/pkg/bridge"
	"go-example/pkg/c"
)

func main() {

	a := &a.A{}

	bridge := bridge.NewBridge(a)

	c := c.C{Bridge: *bridge}

	b := b.B{C: c}

	a.B = b

	a.Print()
	b.Print()
	c.Print()

}
