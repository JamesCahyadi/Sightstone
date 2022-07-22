package client

import (
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/JamesCahyadi/Sightstone/pkg/connect"
	"github.com/gorilla/websocket"
)

type LeagueClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
	Ws       *Websocket
}

type Websocket struct {
	BaseURL string
	Dialer  *websocket.Dialer
}

func (lc *LeagueClient) HttpRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, lc.BaseURL+endpoint, body)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(lc.Username, lc.Password)
	return lc.Client.Do(req)
}

func (lc *LeagueClient) WebSocketRequest() (*websocket.Conn, error) {
	authToken := base64.StdEncoding.EncodeToString([]byte(lc.Username + ":" + lc.Password))
	header := http.Header{}
	header.Add("Authorization", "Basic "+authToken)
	conn, _, err := lc.Ws.Dialer.Dial(lc.Ws.BaseURL, header)
	return conn, err

}

func New() *LeagueClient {
	clientProcess := connect.Try()
	clientInfo := clientProcess.Commandline
	port, password, caCertPool := connect.GetDetails(clientInfo)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	ws := &Websocket{
		BaseURL: "wss://127.0.0.1:" + port,
		Dialer: &websocket.Dialer{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return &LeagueClient{
		BaseURL:  "https://127.0.0.1:" + port,
		Username: "riot",
		Password: password,
		Client:   client,
		Ws:       ws,
	}
}
