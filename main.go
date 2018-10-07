package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

const (
	cat = "kitjs/AuthKey_"

	alg = "ES256"
	iss = "Y82E2K77P5"

	ckPath   = "/database/1/iCloud.telos.atec/development/public/records/modify"
	ckHeader = "X-Apple-CloudKit-Request-"
)

var (
	etc = os.Getenv("ETC")

	client http.Client

	kids = map[string]string{
		"music": "CUG44HA5T5",
		"map":   "YKVC29UG5H",
		"cloud": "",
	}
	keys = make(map[string]*ecdsa.PrivateKey)
)

// loads keys
func init() {

	for svc, kid := range kids {

		path := etc
		if svc == "cloud" {
			path += "cloudkit/eckey.pem"
		} else {
			path += "/" + svc + cat + kid + ".p8"
		}
		fmt.Println("loading key from path: " + path)

		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("reading p8 file:", err)
		}
		block, _ := pem.Decode(b)

		var key *ecdsa.PrivateKey
		if svc == "cloud" {
			key, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				log.Fatal("parsing cloudkit ec key:", err)
			}
		} else {
			tmp, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				log.Fatal("parsing block:", err)
			}
			key = tmp.(*ecdsa.PrivateKey)
		}

		keys[svc] = key
	}
}

func middleware(w http.ResponseWriter, r *http.Request) bool {

	fmt.Println(r.Method + " " + r.URL.Path + " " + r.RemoteAddr + " " + r.Host)

	payload := []byte(`
{
	"operations": [
		{
			"operationType": "create",
			"record": {
				"recordType": "Hit",
				"fields": {
					"method": { "value": "` + r.Method + `" },
					"path": { "value": "` + r.URL.Path + `" },
					"remoteAddr": { "value": "` + r.RemoteAddr + `" },
					"host": { "value": "` + r.Host + `" }
				}
			}
		}
	]
}`)

	date := time.Now().UTC().Format(time.RFC3339)

	h := sha256.New()
	h.Write(payload)
	body := base64.StdEncoding.EncodeToString(h.Sum(nil))

	msg := strings.Join([]string{date, body, ckPath}, ":")
	h = sha256.New()
	h.Write([]byte(msg))

	sig, err := keys["cloud"].Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		fmt.Println("signing msg:", err)
	}
	encodedSig := string(base64.StdEncoding.EncodeToString(sig))
	req, _ := http.NewRequest(http.MethodPost, "https://api.apple-cloudkit.com"+ckPath, bytes.NewBuffer(payload))
	req.Header.Set(ckHeader+"KeyID", "b9f504ff7c0ef5d8b1dc6a1d12e597b3ab5fb9a8e6f24632486c15fb2a8d7f3e")
	req.Header.Set(ckHeader+"ISO8601Date", date)
	req.Header.Set(ckHeader+"SignatureV1", encodedSig)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("doing req:", err)
	}

	b, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(b))

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	return true
}

func hit(w http.ResponseWriter, r *http.Request) {

	if ok := middleware(w, r); !ok {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sign(w http.ResponseWriter, r *http.Request) {

	if ok := middleware(w, r); !ok {
		return
	}

	svc := mux.Vars(r)["svc"]
	if svc != "music" && svc != "map" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jwtToken := &jwt.Token{
		Header: map[string]interface{}{
			"alg": alg,
			"kid": kids[svc],
		},
		Claims: jwt.MapClaims{
			"iss": iss,
			"exp": time.Now().Unix() + 3000,
		},
		Method: jwt.SigningMethodES256,
	}

	bearer, err := jwtToken.SignedString(keys[svc])
	if err != nil {
		fmt.Println("signing bearer:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(bearer))
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", hit)
	r.HandleFunc("/{svc}", sign)

	authEtc := etc + "auth"

	fmt.Println(http.ListenAndServeTLS(":443", authEtc+"/server.crt", authEtc+"/server.key", r))
}
