package main

import (
	"fmt"
	"math"
)

func sayGreeting(n string) {
	fmt.Printf("Good Morning %v \n", n)
}

func sayByeBye(n string) {
	fmt.Printf("Good Bye %v \n", n)
}

func cycleName(n []string, f func(string)) {
	for _, v := range n {
		f(v)
	}
}

func circleArea(r float64) float64 { // Returns a float
	return math.Pi * r * r
}
func main() {

	// Functions

	sayGreeting("Raymond")
	sayByeBye("Reddington")

	cycleName([]string{"Raymond", "Reddington", "Zuma", "Elizabeth", "Tom"}, sayGreeting)

	a1 := circleArea(10)
	a3 := circleArea(15.2)

	fmt.Println(a1, a3)
	fmt.Printf("Circle 1 is %0.1f and Circle 2 is %0.1f \n", a1, a3)
}
