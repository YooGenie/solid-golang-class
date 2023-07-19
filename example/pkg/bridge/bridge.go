package bridge

import (
	"go-example/pkg/printer"
)

type Bridge struct {
	p printer.Printer
}

func NewBridge(p printer.Printer) *Bridge {
	return &Bridge{p: p}
}

func (b Bridge) Print() {
	b.p.Print()
}
