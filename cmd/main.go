package main

import (
	"fmt"

	"github.com/JamesCahyadi/Sightstone/pkg/client"
	"github.com/JamesCahyadi/Sightstone/pkg/friend"
)

func main() {
	lc := client.New()
	fg := "sightstone"

	// should check if the group exists or something
	ok := friend.FindGroup(lc, fg)
	if !ok {
		panic("Couldn't find group" + fg) // should keep listening until the group is created instead of just panicing
	}

	friends := friend.GetFriendsFromGroup(lc, fg)
	fmt.Printf("%+v\n", friends)
}
