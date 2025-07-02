package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type request struct {
	Keywords []string `json:"keywords"`
	Message  string   `json:"message"`
}

type response struct {
	Error   *string `json:"error"`
	Message *string `json:"message"`
}

var addr string

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("missing cli args")
	}
	addr = os.Args[1]
	kws := []string{"abc"}
	status, err := postMessage(kws, "hello world")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(status)
	msg, err := getMessage(kws)
	fmt.Println(msg)
	fmt.Println(*msg.Message)
}

func getMessage(kws []string) (*response, error) {
	uri := fmt.Sprintf("%s/message?keywords=%s", addr, strings.Join(kws, ","))
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := &response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func postMessage(kws []string, msg string) (int, error) {
	var buf bytes.Buffer
	body := request{
		Keywords: kws,
		Message:  msg,
	}
	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return 0, err
	}
	uri := fmt.Sprintf("%s/message", addr)
	resp, err := http.Post(uri, "application/json", &buf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
