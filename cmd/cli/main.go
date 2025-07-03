package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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
	var wg sync.WaitGroup
	ws := 500
	rps, wps := 500, 250
	secs := 10
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(secs))
	defer cancel()
	wg.Add(ws)
	reqs := 0
	for range ws {
		go func() {
			defer wg.Done()
			reqs += work(ctx, rps, wps)
		}()
	}
	wg.Wait()
	fmt.Println(reqs)
}

func getMessage(kws []string) (*response, error) {
	uri := fmt.Sprintf("%s/message?keywords=%s", addr, strings.Join(kws, ","))
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)
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
	defer io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func randWord() string {
	len := 2 + rand.Int()%8
	runes := make([]rune, len)
	for i := range runes {
		runes[i] = rune('a') + rune(rand.Int()%26)
	}
	return string(runes)
}

func randWords() []string {
	len := 2 + rand.Int()%8
	words := make([]string, len)
	for i := range len {
		words[i] = randWord()
	}
	return words
}

func work(ctx context.Context, rps, wps int) int {
	tr := time.NewTicker(time.Millisecond * time.Duration(1000/rps))
	defer tr.Stop()
	tw := time.NewTicker(time.Millisecond * time.Duration(1000/wps))
	defer tw.Stop()
	reqs := 0
	for {
		select {
		case <-tr.C:
			kws := randWords()
			_, err := getMessage(kws)
			reqs++
			if err != nil {
				log.Fatal(err)
			}
		case <-tw.C:
			kws := randWords()
			msg := strings.Join(randWords(), " ")
			status, err := postMessage(kws, msg)
			reqs++
			if err != nil || status != 200 {
				log.Fatal(status, err)
			}
		case <-ctx.Done():
			return reqs
		}
	}
}
