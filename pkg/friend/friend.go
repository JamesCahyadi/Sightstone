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
	GameInfo       GameInfo `json:"lol"`
}

type FriendGroup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// func (f *Friend) GetStatus() string {
// 	// these statuses may be completely wrong
// 	statusMap := map[string]bool{
// 		"inGame": f.Availability == "dnd" && f.GameInfo != (GameInfo{}),

// 		// data when a user is online
// 		// map[data:map[availability:chat displayGroupId:1 displayGroupName:MOBILE gameName:ImStuB gameTag:NA1 groupId:1 groupName:Sightstone icon:744 id:1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net isP2PConversationMuted:false lastSeenOnlineTimestamp:<nil> lol:map[championId: companionId:6028 damageSkinId:1 gameQueueType: gameStatus:outOfGame iconOverride: level:532 mapId: mapSkinId:1 masteryScore:520 profileIcon:744 puuid:1e57e4b9-2bf8-500b-be49-75677751bf18 rankedLeagueDivision:II rankedLeagueQueue:RANKED_FLEX_SR rankedLeagueTier:PLATINUM rankedLosses:0 rankedPrevSeasonDivision:III rankedPrevSeasonTier:PLATINUM rankedSplitRewardLevel:0 rankedWins:39 regalia:{"bannerType":2,"crestType":2,"selectedPrestigeCrest":21} skinVariant: skinname:] name:ImStuB note: patchline:live pid:1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net platformId:NA1 product:league_of_legends productName:League of Legends puuid:1e57e4b9-2bf8-500b-be49-75677751bf18 statusMessage: summary: summonerId:8.3050107e+07 time:1.658368188254e+12] eventType:Update uri:/lol-chat/v1/friends/1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net
// 		// map[data:map[availability:chat displayGroupId:1 displayGroupName:MOBILE gameName:ImStuB gameTag:NA1 groupId:1 groupName:Sightstone icon:744 id:1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net isP2PConversationMuted:false lastSeenOnlineTimestamp:<nil> lol:map[championId: gameQueueType: gameStatus:outOfGame level:532 mapId: masteryScore:520 profileIcon:744 puuid:1e57e4b9-2bf8-500b-be49-75677751bf18 rankedLeagueDivision:II rankedLeagueQueue:RANKED_FLEX_SR rankedLeagueTier:PLATINUM rankedLosses:0 rankedPrevSeasonDivision:III rankedPrevSeasonTier:PLATINUM rankedSplitRewardLevel:0 rankedWins:39 regalia:{"bannerType":2,"crestType":1,"selectedPrestigeCrest":0} skinVariant: skinname:] name:ImStuB note: patchline:live pid:1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net platformId:NA1 product:league_of_legends productName:League of Legends puuid:1e57e4b9-2bf8-500b-be49-75677751bf18 statusMessage: summary: summonerId:8.3050107e+07 time:1.658367454256e+12] eventType:Update uri:/lol-chat/v1/friends/1e57e4b9-2bf8-500b-be49-75677751bf18@na1.pvp.net]

// 		"online": f.Availability == "chat",

// 		"away": f.Availability == "away",

// 		// data when a user goes offline
// 		// map[data:map[availability:offline displayGroupId:1 displayGroupName:OFFLINE gameName:GateKeeper Wyatt gameTag:NA1 groupId:1 groupName:**Default icon:23 id:907d0441-5b7c-522d-abff-94a12d8dda72@na1.pvp.net isP2PConversationMuted:false lastSeenOnlineTimestamp:<nil> lol:map[] name:GateKeeper Wyatt note: patchline: pid:907d0441-5b7c-522d-abff-94a12d8dda72@na1.pvp.net platformId: product: productName: puuid:907d0441-5b7c-522d-abff-94a12d8dda72 statusMessage: summary: summonerId:6.90696e+07 time:0] eventType:Update uri:/lol-chat/v1/friends/907d0441-5b7c-522d-abff-94a12d8dda72@na1.pvp.net]
// 		"offline": f.Availability == "offline",
// 	}

// 	for key := range statusMap {
// 		if statusMap[key] {
// 			return key
// 		}
// 	}

// 	return "offline"
// }

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

func Listen(lc *client.LeagueClient, groupId int) {
	conn, err := lc.WebSocketRequest()
	onlineFriends := make(map[string]bool) // acts like a set
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// this for loop and the one below run simultaneously
	go func() {
		defer close(done)
		for {
			// the server sends a response which gets captured in this message
			_, message, err := conn.ReadMessage()
			if err != nil {
				panic(err)
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
			// so the client sends this to ask if there is any new friend information... subscribe to friend api events from lcu
			err := conn.WriteMessage(websocket.TextMessage, []byte("[5, \"OnJsonApiEvent_lol-chat_v1_friends\"]"))
			if err != nil {
				panic(err)
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
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

	// ignore all events regarding friends outside the group
	if friend.DisplayGroupId != groupId {
		return
	}

	_, isFriendInOnlineFriends := onlineFriends[friend.Name]

	if friend.Availability == "offline" || friend.Availability == "mobile" { // when a friend goes offline, remove them from the list
		delete(onlineFriends, friend.Name)
		// fmt.Println("removing  "+friend.Name+"... current onlineFriends are:", onlineFriends) // send offline toast notification here
		alert.Send(fmt.Sprintf("%s just went offline :(", friend.Name))
	} else if !isFriendInOnlineFriends { // when a user comes online, add them to the list
		onlineFriends[friend.Name] = true
		// fmt.Println("adding "+friend.Name+" ... current onlineFriends are:", onlineFriends)
		alert.Send(fmt.Sprintf("%s is online! Go and invite them to your next game!", friend.Name))
	}

}
