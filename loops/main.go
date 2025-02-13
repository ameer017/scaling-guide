package main

import "fmt"

func main() {
	// x := 0

	// for x < 5 {
	// 	fmt.Println("Value of x is:",x)
	// 	x++
	// }

	// for i := 0; i < 5; i++ {
	// 		fmt.Println("Value of i is:",i)

	// }

	names := []string{"Raymond", "Reddington", "Zuma", "Elizabeth", "Tom"}
	// for n := 0; n < len(names); n++ {
	// 		fmt.Println(names[n])

	// }

	for index, name := range names {
		fmt.Printf("The value at index %v is %v \n", index, name)
	}

}
