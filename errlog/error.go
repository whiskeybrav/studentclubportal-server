package errlog

import (
	"fmt"
	"runtime/debug"
)

func LogError(context string, err error) {
	fmt.Println(context)
	fmt.Println(err)
	fmt.Println(string(debug.Stack()))
}
