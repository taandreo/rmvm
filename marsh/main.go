package main

import (
	"encoding/json"
	"fmt"
)

type test struct {
	N string
	I int
	B bool
}

func main(){
	color()
	t := test{"ok", 42, true}
	fmt.Println("ok")
	fmt.Println(t)
	b, err := json.Marshal(t)
	fmt.Printf("%v\n", err)
	println(string(b))
}

func color() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
}