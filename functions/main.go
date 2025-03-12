package main

import "fmt"

func sayGreeting(n string) {
	fmt.Printf("Good Morning %v \n", n)
}

func sayByeBye(n string) {
	fmt.Printf("Good Bye %v \n", n)
}
func main() {

	// Functions

	sayGreeting("Raymond")
	sayByeBye("Reddington")
}
