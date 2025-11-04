package main

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

type User struct {
	ID     int    `toon:"id"`
	Name   string `toon:"name"`
	Active bool   `toon:"active"`
}

type UserList struct {
	Users []User `toon:"users"`
	Count int    `toon:"count"`
}

func main() {
	doc := UserList{
		Users: []User{
			{ID: 1, Name: "Ada", Active: true},
			{ID: 2, Name: "Bob", Active: false},
		},
		Count: 2,
	}

	encoded, err := toon.Marshal(doc, toon.WithLengthMarkers(true))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))

	var out UserList
	if err := toon.Unmarshal(encoded, &out); err != nil {
		panic(err)
	}
	fmt.Printf("users: %d\n", len(out.Users))
	fmt.Printf("first user: %s\n", out.Users[0].Name)
}
