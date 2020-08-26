package main

import "fmt"

type position struct {
	x float32
	y float32
}

type badGuy struct {
	name   string
	health int
	pos    position
}

func whereIsBadGuyUsingPtr(b *badGuy) {
	x := b.pos.x // no need to dereference here, Go does it for us
	y := b.pos.y
	fmt.Println("{", x, ",", y, ")")
}

func addOne(x int) {
	x = x + 1
}

func addOneUsintPtr(x *int) {
	*x = *x + 1 // dereferencing of the pointer (getting value out of it)
}

func main() {
	x := 5

	xPtr := &x
	var equivalent *int = &x 

	fmt.Println(xPtr)
	fmt.Println(equivalent)

	addOne(x)
	fmt.Println(x)

	addOneUsintPtr(xPtr)
	fmt.Println(x)

	p := position{4, 2}
	b := badGuy{"Jabba The Hut", 100, p}
	whereIsBadGuyUsingPtr(&b)
}