package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/atecce/auth/alert"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

const (
	cat = "kitjs/AuthKey_"

	alg = "ES256"
	iss = "Y82E2K77P5"
)

var (
	etc = os.Getenv("ETC")

	kids = map[string]string{
		"music": "CUG44HA5T5",
		"map":   "YKVC29UG5H",
	}
	keys = make(map[string]*ecdsa.PrivateKey)
)

// loads keys
func init() {

	for svc, kid := range kids {

		path := etc + "/" + svc + cat + kid + ".p8"
		fmt.Println("loading key from path: " + path)

		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("reading p8 file:", err)
		}
		block, _ := pem.Decode(b)

		var key *ecdsa.PrivateKey
		tmp, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal("parsing block:", err)
		}
		key = tmp.(*ecdsa.PrivateKey)

		keys[svc] = key
	}
}

func middleware(w http.ResponseWriter, r *http.Request) bool {

	fmt.Println(r.Method + " " + r.URL.Path + " " + r.RemoteAddr + " " + r.Host)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet || r.Host != "auth.atec.pub" || r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	return true
}

func hit(w http.ResponseWriter, r *http.Request) {

	if ok := middleware(w, r); !ok {
		return
	}

	err := alert.Send(r)
	if err != nil {
		fmt.Println("alerting:", err)
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
