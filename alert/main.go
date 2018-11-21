package alert

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ckPath   = "/database/1/iCloud.telos.atec/development/public/records/modify"
	ckPrefix = "X-Apple-CloudKit-Request-"
)

type alert struct {
	Operations []operation `json:"operations"`
}

type operation struct {
	OperationType string `json:"operationType"`
	Record        record `json:"record"`
}

type record struct {
	RecordType string `json:"recordType"`
	Fields     fields `json:"fields"`
}

type fields struct {
	Method     value `json:"method"`
	Path       value `json:"path"`
	RemoteAddr value `json:"remoteAddr"`
	Host       value `json:"host"`
	Body       value `json:"body"`
}

type value struct {
	Value string `json:"value"`
}

var key *ecdsa.PrivateKey

func init() {
	path := os.Getenv("ETC") + "cloudkit/eckey.pem"
	fmt.Println("loading key from path: " + path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("reading p8 file:", err)
	}
	block, _ := pem.Decode(b)

	key, err = x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Fatal("parsing cloudkit ec key:", err)
	}
}

// Send makes an HTTP request to CloudKit with a signed payload
func Send(r *http.Request) error {

	// DoS vector
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("[ERROR] alert.Send(r): reading body", err)
	}

	payload, err := json.Marshal(alert{
		Operations: []operation{
			operation{
				OperationType: "create",
				Record: record{
					RecordType: "Hit",
					Fields: fields{
						Method:     value{r.Method},
						Path:       value{r.URL.Path},
						RemoteAddr: value{r.RemoteAddr},
						Host:       value{r.Host},
						Body:       value{string(body)},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	date := time.Now().UTC().Format(time.RFC3339)

	h := sha256.New()
	h.Write(payload)
	msg := strings.Join([]string{date, base64.StdEncoding.EncodeToString(h.Sum(nil)), ckPath}, ":")

	h = sha256.New()
	h.Write([]byte(msg))
	sig, err := key.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return err
	}
	encodedSig := string(base64.StdEncoding.EncodeToString(sig))

	req, err := http.NewRequest(http.MethodPost, "https://api.apple-cloudkit.com"+ckPath, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	// TODO make KeyID configurable
	req.Header.Set(ckPrefix+"KeyID", "8b1b74e90bf14d3869f296d39d71fa32a13a879d87c09089005cf72e5966c70a")
	req.Header.Set(ckPrefix+"ISO8601Date", date)
	req.Header.Set(ckPrefix+"SignatureV1", encodedSig)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("doing req:", err)
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("reading body:", err)
		return err
	}
	fmt.Println("[INFO] alert:", string(b))

	return nil
}
