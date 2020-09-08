// apt - abstract picture tree
package apt

import "math"

type Node interface {
	Eval(x, y float32) float32
	String() string
}

// LeafNode has no children
type LeafNode struct{}

// SingleNode has one child
type SingleNode struct {
	Child Node
}

// DoubleNode has two children
type DoubleNode struct {
	LeftChild  Node
	RightChild Node
}

type OpPlus struct {
	// operation plus is a double node (embeds another struct), like in inheritance
	DoubleNode
}

type OpSin struct {
	SingleNode
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Child.Eval(x, y))))
}

func (op *OpSin) String() string {
	return "( Sin " + op.Child.String() + " )"
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) + op.RightChild.Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " +  op.LeftChild.String() + " " + op.RightChild.String() + " )"
}

type OpX LeafNode

func (op *OpX) Eval(x, y float32) float32 {
	return x
}

func (op *OpX) String() string {
	return "X"
}

type OpY LeafNode

func (op *OpY) Eval(x, y float32) float32 {
	return y
}

func (op *OpY) String() string {
	return "Y"
}
