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
	scores = append(scores, 200) // returns a new slice with the new value
	fmt.Println(scores)

	//Slice Ranges ==> [start:end]
	// Declare an array of fixed size 4 with string values
	var names = [4]string{"Al Ameer", "Abdullah", "Raji", "Akintayo"}

	// Create a slice from index 1 to 3 (excluding index 3)
	rangeOne := names[1:3] // This will include "Abdullah" (index 1) and "Raji" (index 2)

	// Create a slice from index 2 to the end of the array
	rangeTwo := names[2:] // This will include "Raji" (index 2) and "Akintayo" (index 3)

	// Print the slices
	fmt.Println(rangeOne) // Output: [Abdullah Raji]
	fmt.Println(rangeTwo) // Output: [Raji Akintayo]

	rangeOne = append(rangeOne, "Alicia") // Add "Alicia" to the slice

	// Print the updated slices
	fmt.Println(rangeOne) // Output: [Abdullah Raji Alicia]
}
