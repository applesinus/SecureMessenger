package main

import (
	"fmt"
	"strings"
)

func enterArgs(amount int) []string {
	args := make([]string, 0)
	switch {
	case amount > 0:
		fmt.Printf("Enter %d args: \n", amount)
		var arg string
		for i := 0; i < amount; i++ {
			fmt.Scan(&arg)
			args = append(args, arg)
		}
	// 0 means that there's undefined amount of arguments
	case amount == 0:
		fmt.Printf("Enter args (to stop enter 'done'): \n")
		var arg string
		for i := 0; true; i++ {
			fmt.Scan(&arg)
			if strings.ToLower(arg) == "done" {
				break
			}
			args = append(args, arg)
		}

	default:
		fmt.Printf("ERROR! Amount of arguments is less than 0 in main.go/enterArgs.\n")
	}
	fmt.Println()
	return args
}

func hr() {
	fmt.Print("\n====================\n")
}

func main() {
	/*str := "kekus"
	newStr := DES.ShuffleIPtest([]byte(str), true, 1)
	fmt.Println("Shuffled: \"", string(newStr), "\"")
	oldStr := DES.ShuffleIPRevtest(newStr, true, 1)
	fmt.Println("Unshuffled: \"", string(oldStr), "\"")*/
}
