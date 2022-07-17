package main

import (
	"fmt"

	"github.com/JamesCahyadi/Sightstone/pkg/client_connect"
)

func main() {
	port, password, caCertPool := client_connect.Connect()
	fmt.Println(port, password, caCertPool)

}
