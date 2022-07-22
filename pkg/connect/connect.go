// handles connecting to the league client
package connect

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

// repeatedly tries to connect to the league client
func Try() *Win32_Process {
	ticker := time.NewTicker(time.Second * 2)
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
			time.Sleep(10 * time.Second) // wait a while for the client to finish loading
			return clientProcess
		}
	}
}

// gets details about the league connection
func GetDetails(clientInfo string) (port string, password string, caCertPool *x509.CertPool) {
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
