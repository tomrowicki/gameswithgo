package main

import "fmt"

func main() {
	/*
		var a int // declaration; defaults to 0
		a = 5     // assignment
		b := 6    // declaration & assignment

		fmt.Println(a)
		fmt.Println(b)
	*/

	// a := 5
	// b := 3.14

	// fmt.Println(a + int(b))

	var a uint8 = 255
	var b uint8 = 10

	fmt.Println(a + b) // wrapping around the max value

	/*
		i := 0
		for i < 10 { // equivalent to while loop
			fmt.Println(i)
			i++
		}
	*/

	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}
