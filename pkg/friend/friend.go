package friend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/JamesCahyadi/Sightstone/pkg/alert"
	"github.com/JamesCahyadi/Sightstone/pkg/client"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

type GameInfo struct {
	GameMode      string `json:"gameMode"`
	GameQueueType string `json:"gameQueueType"`
	GameStatus    string `json:"gameStatus"`
	TimeStamp     string `json:"timeStamp"`
}

type Friend struct {
	Id             string   `json:"id"`
	DisplayGroupId int      `json:"displayGroupId"`
	Availability   string   `json:"availability"`
	Name           string   `json:"name"`
	ProductName    string   `json:"productName"`
	GameInfo       GameInfo `json:"lol"`
}

type FriendGroup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// does an initial scan of friends and adds online friends into the map
// this is to prevent being notified of who is online when opening the client, since you probably will check yourself anyways
func InitialScan(lc *client.LeagueClient, groupId int, onlineFriends map[string]bool) {
	var friends []Friend

	resp, err := lc.HttpRequest(http.MethodGet, "/lol-chat/v1/friends", nil)
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

	for _, friend := range friends {
		if friend.DisplayGroupId == groupId &&
			friend.Availability != "dnd" && // we include dnd here so that when the friend finishes their game, the user will be notified
			friend.Availability != "offline" &&
			friend.Availability != "mobile" {
			onlineFriends[friend.Name] = true
		}
	}
}

func FindGroup(lc *client.LeagueClient, target string) int {
	var friendGroups []FriendGroup
	resp, err := lc.HttpRequest(http.MethodGet, "/lol-chat/v1/friend-groups", nil)
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
			return group.Id
		}
	}

	return -1
}

func Listen(lc *client.LeagueClient, groupId int, onlineFriends map[string]bool) {
	conn, err := lc.WebSocketRequest()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			// the server sends a response which gets captured in this message
			_, message, err := conn.ReadMessage()
			if err != nil {
				close(done)
				return
			}

			processMessage(message, groupId, onlineFriends)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// client sends this to ask if there is any new friend information aka subscribe to friend api events from lcu
			err := conn.WriteMessage(websocket.TextMessage, []byte("[5, \"OnJsonApiEvent_lol-chat_v1_friends\"]"))
			if err != nil {
				panic(err)
			}
		case <-interrupt:
			log.Println("interrupt")

			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}

	}
}

type Message struct {
	Data map[string]Friend `json:"data"`
}

func processMessage(msg []byte, groupId int, onlineFriends map[string]bool) {
	if len(msg) == 0 {
		return
	}

	var msgSlice []interface{}
	err := json.Unmarshal(msg, &msgSlice)
	if err != nil {
		panic(err)
	}

	m := msgSlice[2].(map[string]interface{})

	var friend Friend
	err = mapstructure.Decode(m["data"], &friend)
	if err != nil {
		panic(err)
	}

	// i don't like valorant because it's too hard :(
	if friend.ProductName == "VALORANT" {
		return
	}

	// ignore all events regarding friends outside the group
	if friend.DisplayGroupId != groupId {
		_, isFriendInOnlineFriends := onlineFriends[friend.Name]

		// if we move a friend outside of our group, remove them from onlineFriends just for the sake of ensuring onlineFriends doesn't contain stale data
		if isFriendInOnlineFriends {
			delete(onlineFriends, friend.Name)
		}
		return
	}

	_, isFriendInOnlineFriends := onlineFriends[friend.Name]

	if friend.Availability == "offline" || friend.Availability == "mobile" { // when a friend goes offline, remove them from the list
		delete(onlineFriends, friend.Name)
		alert.Send(fmt.Sprintf("%s just went offline :(", friend.Name))
	} else if friend.Availability == "dnd" { // if a user is in game
		alert.Send(fmt.Sprintf("%s is currently in a game, we'll notify you once they are finished", friend.Name))
	} else if !isFriendInOnlineFriends { // when a user comes online, add them to the list
		onlineFriends[friend.Name] = true
		alert.Send(fmt.Sprintf("%s is online! Go and invite them to your next game!", friend.Name))
	}
}
