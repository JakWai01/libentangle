package main

import (
	"fmt"
	"log"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

var (
	name       = "Jean Done"
	email      = "jean@done.com"
	rsaBits    = 2048
	passphrase = "password"
)

func main() {
	// Generate RSA key
	rsaKey, err := helper.GenerateKey(name, email, []byte(passphrase), "rsa", rsaBits)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rsaKey)

	privKey, err := crypto.NewKeyFromArmored(rsaKey)

	pubKey, err := privKey.GetArmoredPublicKey()

	armor, err := helper.EncryptMessageArmored(pubKey, "Hallo Welt")

	decrypted, err := helper.DecryptMessageArmored(rsaKey, []byte(passphrase), armor)

	fmt.Println(decrypted)

}
