package repositories

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/mattn/go-jsonpointer"
	"github.com/pirosuke/slack-bot-hydration/internal/file"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/models"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/slack"
)

// SlackRepository controls posts to Slack.
type SlackRepository struct {
	Token        string
	ViewsDirPath string
}

// PostHydrationAddResult posts hydration added result message.
func (repo *SlackRepository) PostHydrationAddResult(userName string, channel string, hydration models.Hydration, dailyAmount int64) ([]byte, error) {
	var err error
	var resp []byte
	titleSection := slack.TextSectionBlock{
		Type: "section",
		Text: slack.ContentBlock{
			Type: "mrkdwn",
			Text: "@" + userName + " が飲み物を飲みました\n本日の合計量は " + strconv.FormatInt(dailyAmount, 10) + "ml です",
		},
	}

	contentList := []slack.ContentBlock{
		{
			Type: "mrkdwn",
			Text: "*飲んだもの:*\n" + hydration.Drink,
		},
		{
			Type: "mrkdwn",
			Text: "*摂取量:*\n" + strconv.FormatInt(hydration.Amount, 10) + "ml",
		},
	}

	contentSection := slack.FieldsSectionBlock{
		Type:   "section",
		Fields: contentList,
	}

	var requestParams interface{}

	requestJSON := `{"channel": "", "blocks": []}`
	err = json.Unmarshal([]byte(requestJSON), &requestParams)
	if err != nil {
		return resp, err
	}

	jsonpointer.Set(requestParams, "/channel", channel)
	jsonpointer.Set(requestParams, "/blocks", []interface{}{
		titleSection,
		contentSection,
	})

	requestParamsJSON, err := json.Marshal(requestParams)
	if err != nil {
		return resp, err
	}

	//fmt.Println(string(requestParamsJSON))

	resp, err = slack.PostJSON(repo.Token, "chat.postMessage", string(requestParamsJSON))

	return resp, err
}

// OpenHydrationAddView opens modal for adding Hydration.
func (repo *SlackRepository) OpenHydrationAddView(triggerID string) ([]byte, error) {
	var err error
	var resp []byte
	var requestParams, view interface{}

	viewPath := filepath.Join(repo.ViewsDirPath, "record_form.json")
	if !file.FileExists(viewPath) {
		return resp, fmt.Errorf("View file does not exist: %s", viewPath)
	}

	requestJSON := `{"trigger_id": "", "view": {}}`
	err = json.Unmarshal([]byte(requestJSON), &requestParams)
	if err != nil {
		return resp, err
	}

	viewJSON, err := ioutil.ReadFile(viewPath)
	if err != nil {
		return resp, err
	}

	err = json.Unmarshal([]byte(viewJSON), &view)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/view", view)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/trigger_id", triggerID)
	if err != nil {
		return resp, err
	}

	requestParamsJSON, err := json.Marshal(requestParams)
	if err != nil {
		return resp, err
	}

	//fmt.Println(string(requestParamsJSON))
	resp, err = slack.PostJSON(repo.Token, "views.open", string(requestParamsJSON))
	//fmt.Println(string(resp))

	return resp, err
}
