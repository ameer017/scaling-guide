package main

import "fmt"

func main() {

	// Boolean
	age := 45

	fmt.Println(age >= 30)
	fmt.Println(age <= 30)
	fmt.Println(age == 45)
	fmt.Println(age != 30)

	if age < 30 {
		fmt.Println("Age is less than 30")
	} else if age < 40 {
		fmt.Println("Age is less than 40")
	} else {
		fmt.Println("Age is greater than 40")
	}

	names := []string{"Raymond", "Reddington", "Zuma", "Elizabeth", "Tom"}

	for index, value := range names {
		if index == 1 {
			fmt.Println("Continuing at pos", index)
			continue
		}

		if index > 2 {
			fmt.Println("Breaking at pos", index);
			break;
		}

		fmt.Printf("The value at pos %v is %v \n", index, value)
	}
}
