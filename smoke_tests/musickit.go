package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

func get(client *http.Client, req *http.Request) string {

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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", "https://35.237.184.72/music", nil)
	if err != nil {
		fmt.Println(err)
	}

	res := get(client, req)

	req, err = http.NewRequest("GET", "https://api.music.apple.com/v1/catalog/us/songs/203709340", nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer "+res)

	println(get(client, req))
}
