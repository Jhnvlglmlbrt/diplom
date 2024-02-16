package main

import (
	"crypto/tls"
	"fmt"
	"log"
)

func checkSSL(domain string) {
	config := &tls.Config{}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", config))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
}

func main() {

}
