package main

import "fmt"

func main() {
	// String === variable initialization with var
	var name string = "Al Ameer"
	var lastName = "Abdullah"
	var thirdName string

	fmt.Println(name, lastName, thirdName)

	// string updates
	name = "Al Ameer Abdullah"
	fmt.Println(name)

	// Shorthand

	Surname := "Raji" // works only inside of a function
	fmt.Println(Surname)
}
