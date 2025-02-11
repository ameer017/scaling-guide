package main

// https://pkg.go.dev/std

import (
	"fmt"
	"sort"
	"strings"
)

func main() {

	// Strings method
	var greeting = "Hello There friends!"

	fmt.Println(strings.Contains(greeting, "There"))         // returns a boolean value [ttrue/false]
	fmt.Println(strings.ReplaceAll(greeting, "Hello", "Hi")) // Replaces a specified string with another
	fmt.Println(strings.ToUpper(greeting))                   // converts all letters to uppercase
	fmt.Println(strings.Index(greeting, "ll"))               // returns the index of the first occurrence
	fmt.Println(strings.Split(greeting, " "))                // splits a string into a slice

	// Sort method
	ages := []int{20, 40, 10, 50, 70, 80, 30, 60}

	sort.Ints(ages) // sorts the slice
	fmt.Println(ages)

	index := sort.SearchInts(ages, 30)    // returns the index of the first element >= 30
	fmt.Println("index returned:", index) // If a number is not found, the index returned will be len + 1

	name := []string{"Raymond", "Reddington", "Zuma", "Elizabeth", "Tom"}
	sort.Strings(name) // sorts the slice
	fmt.Println(name)

	fmt.Println(sort.SearchStrings(name, "Tom"))
}
