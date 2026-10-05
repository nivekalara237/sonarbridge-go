package debug

import "fmt"

func PrintLn(ss ...any) {
	fmt.Println("-------------------------------------------------------------")
	for _, s := range ss {
		fmt.Println(s)
	}
	fmt.Println("-------------------------------------------------------------")
}
