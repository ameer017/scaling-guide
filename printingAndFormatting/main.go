package main

import "fmt"

func main() {
	// Print
	fmt.Print("hello, ")
	fmt.Print("World! \n")
	fmt.Print("New Line")

	// Println
	fmt.Println("Hello Ninjas!")
	fmt.Println("Goodbye Ninjas!")

	// `Print` combines all text into a single line unless you include `\n` for line breaks, whereas `Println` automatically prints each statement on a new line.

	age := 35
	name := "Ninja"

	fmt.Println("My name is", name, "and I am", age, "years old")

	// Printf (formatted strings)
	fmt.Printf("My name is %v and I am %v years old \n", name, age) // %v is a placeholder ==> format specifier [v === variable]
	fmt.Printf("My name is %q and I am %v years old \n", name, age) // [q === add quotes to the variables] to be used on strings only
	fmt.Printf("Age is of type %T \n", age)                         // %T gets us the type of the variable
	fmt.Printf("You scored %f poinst! \n", 225.55)                  // to return a floating value
	fmt.Printf("You scored %0.1f poinst! \n", 225.55)               // Specifies the numbers of digits after the decimal [rounded figure!!]

	// Sprintf (save formatted strings)
	s := fmt.Sprintf("My name is %v and I am %v years old \n", name, age)
	fmt.Println(s)
}
