package main

import (
	"fmt"
	"reflect"
)

func main() {
	a := 123
	b := "123"

	fmt.Println(reflect.ValueOf(a).Type())
	fmt.Println(reflect.ValueOf(b).Type())
}
