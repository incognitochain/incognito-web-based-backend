package slacknoti

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var notiChan chan string
var notiArray []string
var notiLock sync.Mutex

func send(text string) {
	content := struct {
		Text string `json:"text"`
	}{
		Text: text,
	}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		log.Println(err)
		return
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Post(os.Getenv("SLACK_MONITOR"), "application/json", bytes.NewReader(contentBytes))
	if resp.Status != "200" || err != nil {
		log.Println(err)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
	}
	defer resp.Body.Close()
}

func StartSlackHook() {
	notiChan = make(chan string)
	t := time.NewTicker(30 * time.Second)
	for {
		select {
		case noti := <-notiChan:
			notiLock.Lock()
			notiArray = append(notiArray, noti)
			notiLock.Unlock()
		case <-t.C:
			notiLock.Lock()
			if len(notiArray) > 0 {
				texts := strings.Join(notiArray, "\n")
				send(texts)
				notiArray = []string{}
			}
			notiLock.Unlock()
		}
	}
}

func SendSlackNoti(msg string) {
	notiChan <- msg
}
