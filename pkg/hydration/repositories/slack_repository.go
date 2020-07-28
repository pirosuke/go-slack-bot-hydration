package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
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

// ShowAlert opens modal for alert.
func (repo *SlackRepository) ShowAlert(triggerID string, title string, text string) ([]byte, error) {

	viewParams := map[string]string{
		"title": title,
		"text":  text,
	}

	var resp []byte

	viewPath := filepath.Join(repo.ViewsDirPath, "alert_dialog.json")
	if !file.FileExists(viewPath) {
		return resp, fmt.Errorf("View file does not exist: %s", viewPath)
	}

	return repo.openView(triggerID, viewPath, viewParams)
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

	resp, err = slack.PostJSON(repo.Token, "chat.postMessage", "application/json", string(requestParamsJSON))
	fmt.Println(string(resp))

	return resp, err
}

// PostHydrationUpdateResult posts hydration added result message.
func (repo *SlackRepository) PostHydrationUpdateResult(userName string, channel string, ts string, hydration models.Hydration, dailyAmount int64) ([]byte, error) {
	var err error
	var resp []byte
	var requestParams, view interface{}

	viewPath := filepath.Join(repo.ViewsDirPath, "result_message.json")
	if !file.FileExists(viewPath) {
		return resp, fmt.Errorf("View file does not exist: %s", viewPath)
	}

	requestJSON := `{"channel": "", "ts": "", "blocks": []}`
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

	err = jsonpointer.Set(requestParams, "/ts", ts)
	if err != nil {
		return resp, err
	}

	requestParamsJSON, err := json.Marshal(requestParams)
	if err != nil {
		return resp, err
	}

	fmt.Println(string(requestParamsJSON))

	resp, err = slack.PostJSON(repo.Token, "chat.update", "application/json", string(requestParamsJSON))
	fmt.Println(string(resp))

	return resp, err
}

// OpenHydrationAddView opens modal for adding Hydration.
func (repo *SlackRepository) OpenHydrationAddView(triggerID string) ([]byte, error) {

	viewParams := map[string]string{
		"callbackID":    "hydration__record_form",
		"metadata":      "",
		"initialDrink":  "",
		"initialAmount": "100",
	}

	return repo.openHydrationEditView(triggerID, viewParams)
}

// OpenHydrationUpdateView opens modal for updating Hydration.
func (repo *SlackRepository) OpenHydrationUpdateView(triggerID string, channel string, ts string, hydration models.Hydration) ([]byte, error) {

	viewParams := map[string]string{
		"callbackID":    "hydration__update_form",
		"metadata":      channel + "-" + ts + "-" + strconv.FormatInt(hydration.ID, 10),
		"initialDrink":  hydration.Drink,
		"initialAmount": strconv.FormatInt(hydration.Amount, 10),
	}

	return repo.openHydrationEditView(triggerID, viewParams)
}

// openHydrationEditView opens modal for adding Hydration.
func (repo *SlackRepository) openHydrationEditView(triggerID string, viewParams map[string]string) ([]byte, error) {
	var resp []byte

	viewPath := filepath.Join(repo.ViewsDirPath, "record_form.json")
	if !file.FileExists(viewPath) {
		return resp, fmt.Errorf("View file does not exist: %s", viewPath)
	}

	return repo.openView(triggerID, viewPath, viewParams)
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
	resp, err = slack.PostJSON(repo.Token, "chat.delete", "application/json", string(requestParamsJSON))
	//fmt.Println(string(resp))

	return resp, err

}

// UploadFile uploads file.
func (repo *SlackRepository) UploadFile(channel string, fileName string, fileType string, filePath string, message string) ([]byte, error) {
	var err error
	var resp []byte

	fileReader, err := os.Open(filePath)
	if err != nil {
		return resp, err
	}
	defer fileReader.Close()

	var requestParams bytes.Buffer
	var fw io.Writer

	writer := multipart.NewWriter(&requestParams)

	fw, _ = writer.CreateFormField("token")
	io.Copy(fw, strings.NewReader(repo.Token))

	fw, _ = writer.CreateFormField("channels")
	io.Copy(fw, strings.NewReader(channel))

	fw, _ = writer.CreateFormField("filename")
	io.Copy(fw, strings.NewReader(fileName))

	fw, _ = writer.CreateFormField("filetype")
	io.Copy(fw, strings.NewReader(fileType))

	fw, _ = writer.CreateFormFile("file", fileReader.Name())
	io.Copy(fw, fileReader)

	if len(message) > 0 {
		fw, _ = writer.CreateFormField("initial_comment")
		io.Copy(fw, strings.NewReader(message))
	}

	writer.Close()

	//fmt.Println(string(requestParamsJSON))
	resp, err = slack.PostBuffer(repo.Token, "files.upload", writer.FormDataContentType(), &requestParams)
	//fmt.Println(string(resp))

	return resp, err

}

// openView opens modal from template.
func (repo *SlackRepository) openView(triggerID string, viewPath string, viewParams map[string]string) ([]byte, error) {
	var err error
	var resp []byte
	var requestParams, view interface{}

	requestJSON := `{"trigger_id": "", "view": {}}`
	err = json.Unmarshal([]byte(requestJSON), &requestParams)
	if err != nil {
		return resp, err
	}

	viewJSONTemplate, err := ioutil.ReadFile(viewPath)
	if err != nil {
		return resp, err
	}

	viewJSON := replaceViewTemplateParams(string(viewJSONTemplate), viewParams)
	//fmt.Println(viewJSON)

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
	resp, err = slack.PostJSON(repo.Token, "views.open", "application/json", string(requestParamsJSON))
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
