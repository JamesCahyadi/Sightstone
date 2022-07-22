package main

import (
	"github.com/JamesCahyadi/Sightstone/pkg/client"
	"github.com/JamesCahyadi/Sightstone/pkg/friend"
)

func main() {
	lc := client.New()
	fg := "sightstone"

	// should check if the group exists or something
	groupId := friend.FindGroup(lc, fg)
	if groupId == -1 {
		groupId = 0 // the default General group
	}

	onlineFriends := make(map[string]bool) // acts like a set, value isn't used
	friend.InitialScan(lc, groupId, onlineFriends)
	friend.Listen(lc, groupId, onlineFriends)
}
