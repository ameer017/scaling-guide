package main

import "fmt"

func main() {

	// Arrays
	var ages [3]int = [3]int{10, 20, 30} // hardcoded array declaration

	// Shorthand
	ages2 := [3]int{10, 20, 30} // shorthand array declaration

	fmt.Println(ages, ages2)
	fmt.Println(len(ages), len(ages2))

	//slices [uses array under the hood]
	var scores = []int{100, 150, 120}
	scores[2] = 110
	scores =append(scores, 200) // returns a new slice with the new value
	fmt.Println(scores)
}
