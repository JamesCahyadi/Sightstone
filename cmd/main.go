package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/JamesCahyadi/Sightstone/pkg/client_connect"
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

func main() {
	port, password, caCertPool := client_connect.Connect()
	baseURL := "https://127.0.0.1:" + port

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	var friends []Friend

	req, err := http.NewRequest(http.MethodGet, baseURL+"/lol-chat/v1/friends", nil)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth("riot", password)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(body, &friends)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(friends)

}
