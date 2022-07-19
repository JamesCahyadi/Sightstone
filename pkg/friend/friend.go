package friend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/JamesCahyadi/Sightstone/pkg/client"
)

type GameInfo struct {
	GameMode      string `json:"gameMode"`
	GameQueueType string `json:"gameQueueType"`
	GameStatus    string `json:"gameStatus"`
	TimeStamp     string `json:"timeStamp"`
}

type Friend struct {
	Id           string   `json:"id"`
	GroupName    string   `json:"groupName"`
	Availability string   `json:"availability"`
	Name         string   `json:"name"`
	GameInfo     GameInfo `json:"lol"`
}

type FriendGroup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func FindGroup(lc *client.LeagueClient, target string) bool {
	var friendGroups []FriendGroup
	resp, err := lc.Do(http.MethodGet, "/lol-chat/v1/friend-groups", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &friendGroups)
	if err != nil {
		panic(err)
	}

	for _, group := range friendGroups {
		if target == strings.ToLower(group.Name) {
			return true
		}
	}

	return false
}

func GetFriendsFromGroup(lc *client.LeagueClient, target string) []Friend {
	var friends []Friend
	// loop through all friends since lol-chat/v1/friend-groups/{id}/friend endpoint is not implemented
	resp, err := lc.Do(http.MethodGet, "/lol-chat/v1/friends", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &friends)
	if err != nil {
		panic(err)
	}

	var groupFriends []Friend
	for _, friend := range friends {
		if strings.ToLower(friend.GroupName) == target {
			groupFriends = append(groupFriends, friend)
		}
	}
	return groupFriends
}
