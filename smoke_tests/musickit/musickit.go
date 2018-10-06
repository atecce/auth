package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
)

func get(req *http.Request) string {

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	return string(b)
}

func newReq(url string) *http.Request {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}

	return req
}

func main() {

	req := newReq("https://api.music.apple.com/v1/catalog/us/songs/203709340")
	req.Header.Add("Authorization", "Bearer "+get(newReq("https://auth.atec.pub/music")))

	println(get(req))
}
