package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

// addr := "https://$(gcloud compute addresses describe auth --region us-east1 --format='value(address)')"
// echo $ADDR

// BEARER=$(curl -k $ADDR/music)
// echo $BEARER

// curl -v -H "Authorization: Bearer $BEARER" "https://api.music.apple.com/v1/catalog/us/songs/203709340"

func main() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", "https://35.237.184.72/music", nil)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	req, _ = http.NewRequest("GET", "https://api.music.apple.com/v1/catalog/us/songs/203709340", nil)
	req.Header.Add("Authorization", "Bearer "+string(b))

	res, _ = client.Do(req)
	b, _ = ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	println(string(b))
}
