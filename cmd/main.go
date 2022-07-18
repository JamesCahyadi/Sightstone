package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	Availability string   `json:"availability"`
	Name         string   `json:"name"`
	GameInfo     GameInfo `json:"lol"`
}

type LeagueClientRequest struct {
	Port       string
	Password   string
	CaCertPool *x509.CertPool
	BaseURL    string
}

func main() {
	lc := client.LeagueClientConnect()

	var friends []Friend

	resp, err := lc.Do(http.MethodGet, "/lol-chat/v1/friends", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(string(body))
	err = json.Unmarshal(body, &friends)
	if err != nil {
		log.Fatal(err)
	}

}
