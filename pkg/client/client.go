package client

import (
	"crypto/tls"
	_ "embed"
	"io"
	"net/http"

	"github.com/JamesCahyadi/Sightstone/pkg/connect"
)

type LeagueClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

func (lc *LeagueClient) Do(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, lc.BaseURL+endpoint, body)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(lc.Username, lc.Password)
	return lc.Client.Do(req)
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

	return &LeagueClient{
		BaseURL:  "https://127.0.0.1:" + port,
		Username: "riot",
		Password: password,
		Client:   client,
	}
}
