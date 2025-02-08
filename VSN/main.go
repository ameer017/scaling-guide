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

	// Ints
	var age int = 20
	var age2 = 22
	var age3 int

	fmt.Println(age, age2, age3)

	// Shorthand
	Age := 20
	fmt.Println(Age)

	// bits and Memory  visit ==> https://go.ord/pkg/builtin for guide
	var num1 int8 = 25
	var num2 = -128
	var num3 uint8 = 255


	fmt.Println(num1, num2, num3)
	Num := 25
	fmt.Println(Num)

	// Floats ==> either 32/64, range doesn't matter but 64 has higher precision compared to 32
	var float1 float32 = 25.5
	var float2 = -128.5
	float3 := 255.5

	fmt.Println(float1, float2, float3)
}
