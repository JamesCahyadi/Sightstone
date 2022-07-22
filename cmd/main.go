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
		panic("Couldn't find group" + fg) // should keep listening until the group is created instead of just panicing
	}

	onlineFriends := make(map[string]bool) // acts like a set, value isn't used
	friend.InitialScan(lc, groupId, onlineFriends)
	// friends := friend.GetFriendsFromGroup(lc, groupId)
	friend.Listen(lc, groupId, onlineFriends)
}
