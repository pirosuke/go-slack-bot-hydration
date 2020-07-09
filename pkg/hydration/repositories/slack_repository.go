package repositories

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

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
	var requestParams, view interface{}

	viewPath := filepath.Join(repo.ViewsDirPath, "result_message.json")
	if !file.FileExists(viewPath) {
		return resp, fmt.Errorf("View file does not exist: %s", viewPath)
	}

	requestJSON := `{"channel": "", "blocks": []}`
	err = json.Unmarshal([]byte(requestJSON), &requestParams)
	if err != nil {
		return resp, err
	}

	viewJSONTemplate, err := ioutil.ReadFile(viewPath)
	if err != nil {
		return resp, err
	}

	viewParams := map[string]string{
		"hydrationID": strconv.FormatInt(hydration.ID, 10),
		"userName":    hydration.Username,
		"drink":       hydration.Drink,
		"amount":      strconv.FormatInt(hydration.Amount, 10),
		"dailyAmount": strconv.FormatInt(dailyAmount, 10),
	}

	viewJSON := replaceViewTemplateParams(string(viewJSONTemplate), viewParams)

	err = json.Unmarshal([]byte(viewJSON), &view)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/blocks", view)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/channel", channel)
	if err != nil {
		return resp, err
	}

	requestParamsJSON, err := json.Marshal(requestParams)
	if err != nil {
		return resp, err
	}

	fmt.Println(string(requestParamsJSON))

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

// DeleteMessage deletes message.
func (repo *SlackRepository) DeleteMessage(channel string, ts string) ([]byte, error) {
	var err error
	var resp []byte
	var requestParams interface{}

	requestJSON := `{"channel": "", "ts": ""}`
	err = json.Unmarshal([]byte(requestJSON), &requestParams)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/channel", channel)
	if err != nil {
		return resp, err
	}

	err = jsonpointer.Set(requestParams, "/ts", ts)
	if err != nil {
		return resp, err
	}

	requestParamsJSON, err := json.Marshal(requestParams)
	if err != nil {
		return resp, err
	}

	//fmt.Println(string(requestParamsJSON))
	resp, err = slack.PostJSON(repo.Token, "chat.delete", string(requestParamsJSON))
	//fmt.Println(string(resp))

	return resp, err

}

func replaceViewTemplateParams(srcText string, params map[string]string) string {
	destText := srcText

	for k, v := range params {
		destText = strings.ReplaceAll(destText, "{{"+k+"}}", v)
	}

	return destText
}
