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

func main() {

	req, err := http.NewRequest(http.MethodGet, "https://35.237.184.72/music", nil)
	if err != nil {
		fmt.Println(err)
	}

	res := get(req)

	req, err = http.NewRequest(http.MethodGet, "https://api.music.apple.com/v1/catalog/us/songs/203709340", nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer "+res)

	println(get(req))
}
