package main

import "fmt"

func sayHello(name string) {
	fmt.Println("Hello,", name)
}

func addOne(x int) int {
	return x + 1
}

func main() {
	sayHello("Bob")
	sayHello("Tom")

	fmt.Println(addOne(11))
}
