package main

import (
	"fmt"
	"log"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

var (
	name    = "Jean Done"
	email   = "jean@done.com"
	rsaBits = 2048
)

func main() {
	fmt.Println(crypto.GetUnixTime())

	// Generate RSA key
	rsaKey, err := crypto.GenerateKey(name, email, "rsa", rsaBits)
	if err != nil {
		log.Fatal(err)
	}

	if key, err := rsaKey.GetArmoredPublicKey(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(key))
	}
}
