package toon_test

import (
	"fmt"

	"github.com/toon-format/toon-go"
)

type User struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Role string `toon:"role"`
}

type Payload struct {
	Users []User `toon:"users"`
}

func Example() {
	in := Payload{Users: []User{{1, "Ada", "admin"}, {2, "Bob", "user"}}}

	encoded, err := toon.MarshalString(in)
	if err != nil {
		panic(err)
	}
	fmt.Println(encoded)

	var out Payload
	if err := toon.UnmarshalString(encoded, &out); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", out)
	// Output:
	// users[2]{id,name,role}:
	//   1,Ada,admin
	//   2,Bob,user
	// {Users:[{ID:1 Name:Ada Role:admin} {ID:2 Name:Bob Role:user}]}
}
