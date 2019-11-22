package errlog

import "fmt"

func LogError(context string, err error) {
	fmt.Println(context)
	fmt.Println(err)
}
