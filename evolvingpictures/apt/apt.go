// apt - abstract picture tree
package apt

type Node interface {
	Eval(x, y float32) float32
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
	// operation plus is a double node (embeds another struct)
	 DoubleNode
}

func (op *OpPlus) Eval(x,y float32) float32 {
	return op.LeftChild.Eval(x,y) + op.RightChild.Eval(x,y)
}

type OpX LeafNode

func (op *OpX) Eval(x,y float32) float32 {
	return x
}

type OpY LeafNode

func (op *OpY) Eval(x,y float32) float32 {
	return y
}