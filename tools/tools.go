package tools

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"log"
	"os"
)

func GenerateTokenKey() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	if err != nil {
		log.Print(err)
	}

	keyFile, err := os.OpenFile(constants.TOKEN_KEY_FILE, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print(err)
		return
	}

	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		log.Print(err)
		os.Exit(2)
	}

	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyFile.Close()
}
