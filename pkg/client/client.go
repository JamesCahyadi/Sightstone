package client

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/StackExchange/wmi"
)

//go:embed riotgames.pem
var caCert []byte

type Win32_Process struct {
	Commandline string
}

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

func LeagueClientConnect() *LeagueClient {
	clientProcess := tryConnect()
	clientInfo := clientProcess.Commandline
	port, password, caCertPool := getConnectionDetails(clientInfo)

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

func findLeagueClientProcess() *Win32_Process {
	var dst []Win32_Process
	q := wmi.CreateQuery(&dst, "WHERE name='LeagueClientUx.exe'")
	if err := wmi.Query(q, &dst); err != nil {
		panic(err)
	}
	if len(dst) == 0 {
		return nil
	} else if len(dst) != 1 {
		panic("found more than one league client open")
	}
	fmt.Println("found process", dst[0])
	return &(dst[0])
}

func tryConnect() *Win32_Process {
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	var clientProcess *Win32_Process

	for {
		select {
		case <-ticker.C:
			clientProcess = findLeagueClientProcess()
			if clientProcess != nil {
				close(quit)
			}
		case <-quit:
			ticker.Stop()
			return clientProcess
		}
	}
}

func getConnectionDetails(clientInfo string) (port string, password string, caCertPool *x509.CertPool) {
	portRegex, err := regexp.Compile("--app-port=([0-9]*)")
	if err != nil {
		panic("invalid port regex")
	}
	portRes := portRegex.FindStringSubmatch(clientInfo)
	if len(portRes) != 2 {
		panic("cannot find port")
	}
	port = portRes[1]

	// obtain password
	passwordRegex, err := regexp.Compile("--remoting-auth-token=([a-zA-Z0-9_-]*)")
	if err != nil {
		panic("invalid remoting auth token regex")
	}
	passwordRes := passwordRegex.FindStringSubmatch(clientInfo)
	if len(passwordRes) != 2 {
		panic("cannot find remoting auth token")
	}
	password = passwordRes[1]

	if err != nil {
		log.Fatal(err)
	}
	caCertPool = x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		panic("cannot add certificate")
	}

	return port, password, caCertPool
}
