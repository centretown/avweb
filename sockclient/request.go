package sockclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	HOST_MELON   = "http://melon:8123"
	API          = "/api/"
	GET          = "GET"
	POST         = "POST"
	CONTENT      = "Content-Type"
	CONTENT_JSON = "application/json"
)

func ClientGet(cmd string) (buf []byte, err error) {

	log.Println(GET, cmd)
	req, err := http.NewRequest(GET, HOST_MELON+API+cmd, nil)
	if err != nil {
		log.Println(GET, err)
		return
	}

	buf, err = ClientRequest(req)
	fmt.Println(string(buf))
	return
}

func ClientPost(cmd string, body string) ([]byte, error) {
	log.Println(POST, cmd, body)
	buf := bytes.NewBuffer(([]byte)(body))

	req, err := http.NewRequest(POST, HOST_MELON+API+cmd, buf)
	if err != nil {
		log.Println(POST, err)
		return buf.Bytes(), err
	}

	req.Header.Add(CONTENT, CONTENT_JSON)
	return ClientRequest(req)
}

func ClientRequest(req *http.Request) (buf []byte, err error) {

	req.Header.Add("Authorization", "Bearer "+Token)
	var (
		resp   *http.Response
		client = &http.Client{}
	)

	resp, err = client.Do(req)
	if err != nil {
		log.Println("client.Do", err)
		return
	}

	buf, err = io.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		log.Println("ReadAll", err)
		return
	}

	return
}

func ClientAuth() (err error) {
	return
}
