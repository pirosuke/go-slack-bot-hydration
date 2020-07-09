package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
Message describes the format for posting message.
*/
type (
	// Config describes config for slack.
	Config struct {
		Token string `json:"token"`
	}

	Message struct {
		Channel string `json:"channel"`
		Text    string `json:"text"`
		Blocks  string `json:"blocks"`
	}
)

/*
PostMessage posts message to slack.
*/
func PostMessage(token string, message Message) ([]byte, error) {

	postMessageJSON, _ := json.Marshal(message)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", strings.NewReader(string(postMessageJSON)))
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

/*
PostJSON posts request to slack.
*/
func PostJSON(token string, command string, paramJSON string) ([]byte, error) {

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://slack.com/api/"+command, strings.NewReader(paramJSON))
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
