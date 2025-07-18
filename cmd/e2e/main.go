package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"maps"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
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

var addr string = "http://localhost:8080"

func loadTest() {
	ws, secs := 10, 1
	rps, wps := 100, 20
	ch := make(chan map[string]string, ws)
	defer close(ch)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(secs))
	defer cancel()
	for range ws {
		go func() {
			ch <- work(ctx, rps, wps)
		}()
	}
	sent := make(map[string]string, secs*wps)
	for range ws {
		maps.Copy(sent, <-ch)
	}
	slog.Info("total requests handled", "reqs", len(sent))
	tests := 10
	for kws, msg := range sent {
		resp, err := getMessage(strings.Split(kws, ","))
		if err != nil {
			log.Fatalf("error during load test %v", err)
		}
		if msg != *resp.Message {
			log.Fatalf("load test failed: %s != %s", msg, *resp.Message)
		}
		tests--
		if tests == 0 {
			break
		}
	}
	slog.Info("load test passed")
}

func notFoundTest() {
	kws := []string{"this", "will", "not", "be", "there"}
	resp, err := getMessage(kws)
	if err != nil {
		log.Fatalf("failed not found test %v", err)
	}
	if resp.Message != nil || *resp.Error != "not found" {
		log.Fatalf("failed not found test: value exists")
	}
	slog.Info("404 test passed")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("missing cli args")
	}
	path := os.Args[1]
	cmd := exec.Command(path)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("could not start the service %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	tests := []func(){loadTest, notFoundTest}
	var wg sync.WaitGroup
	wg.Add(len(tests))
	for _, test := range tests {
		go func() {
			defer wg.Done()
			test()
		}()
	}
	wg.Wait()

	err = cmd.Process.Kill()
	if err != nil {
		log.Fatalf("failed to kill process %v", err)
	}
	cmd.Process.Wait()
	slog.Info("Tests passed...")
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

func work(ctx context.Context, rps, wps int) map[string]string {
	sent := make(map[string]string, 1000)
	tr := time.NewTicker(time.Millisecond * time.Duration(1000/rps))
	defer tr.Stop()
	tw := time.NewTicker(time.Millisecond * time.Duration(1000/wps))
	defer tw.Stop()
	for {
		select {
		case <-tr.C:
			kws := randWords()
			_, err := getMessage(kws)
			if err != nil {
				log.Fatal(err)
			}
		case <-tw.C:
			kws := randWords()
			msg := strings.Join(randWords(), " ")
			status, err := postMessage(kws, msg)
			sent[strings.Join(kws, ",")] = msg
			if err != nil || status != 200 {
				log.Fatal(status, err)
			}
		case <-ctx.Done():
			return sent
		}
	}
}
