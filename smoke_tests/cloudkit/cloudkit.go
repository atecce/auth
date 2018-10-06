package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// smoke payload
var payload = []byte(`
{
	"operations": [
		{
			"operationType": "create",
			"record": {
				"recordType": "Test",
				"fields": {
					"ping": { "value": "pong" }
				}
			}
		}
	]
}`)

func main() {

	// initialize date
	date := time.Now().UTC().Format(time.RFC3339)

	// initialize path
	path := "/database/1/iCloud.telos.atec/development/public/records/modify"

	// encode payload
	h := sha256.New()
	h.Write(payload)
	body := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// join msg
	msg := strings.Join([]string{date, body, path}, ":")

	// encode msg
	h = sha256.New()
	h.Write([]byte(msg))

	// read pem
	b, err := ioutil.ReadFile("/keybase/private/atec/etc/cloudkit/eckey.pem")
	if err != nil {
		log.Fatal(err)
	}

	// decode pem
	block, _ := pem.Decode(b)

	// parse ec private key
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	// sign msg
	sig, err := priv.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		log.Fatal(err)
	}

	// encode sig
	encodedSig := string(base64.StdEncoding.EncodeToString(sig))

	// construct req
	req, _ := http.NewRequest(http.MethodPost, "https://api.apple-cloudkit.com"+path, bytes.NewBuffer(payload))
	req.Header.Set("X-Apple-CloudKit-Request-KeyID", "b9f504ff7c0ef5d8b1dc6a1d12e597b3ab5fb9a8e6f24632486c15fb2a8d7f3e")
	req.Header.Set("X-Apple-CloudKit-Request-ISO8601Date", date)
	req.Header.Set("X-Apple-CloudKit-Request-SignatureV1", encodedSig)

	fmt.Println(req.Header)
	println()

	// do req
	client := http.Client{}
	res, _ := client.Do(req)

	// get res
	b, _ = ioutil.ReadAll(res.Body)
	fmt.Println(string(b))
}
