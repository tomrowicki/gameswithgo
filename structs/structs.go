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

func whereIsBadGuy(b badGuy) {
	x := b.pos.x
	y := b.pos.y
	fmt.Println("{", x, ",", y, ")")
}

func main() {
	var p position
	fmt.Println(p)

	p.x = 5
	p.y = 7
	fmt.Println(p)

	p2 := position{4, 2}
	fmt.Println(p2)

	b := badGuy{"Jabba The Hut", 100, p}
	fmt.Println(b)

	whereIsBadGuy(b)
}
