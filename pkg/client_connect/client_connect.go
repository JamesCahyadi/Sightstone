package client_connect

import (
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/StackExchange/wmi"
)

//go:embed riotgames.pem
var caCert []byte

type Win32_Process struct {
	Commandline string
}

func Connect() (port string, password string, caCertPool *x509.CertPool) {
	clientProcess := tryConnect()
	clientInfo := clientProcess.Commandline
	return getConnectionDetails(clientInfo)
	// obtain port

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
			fmt.Println("retry connect..")
			if clientProcess != nil {
				close(quit)
			}
		case <-quit:
			ticker.Stop()
			fmt.Println("connected!")
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
