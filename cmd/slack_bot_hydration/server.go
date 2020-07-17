package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	jsonpointer "github.com/mattn/go-jsonpointer"
	"github.com/pirosuke/slack-bot-hydration/internal/file"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/config"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/interfaces"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/models"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/repositories"
)

var (
	repo interfaces.HydrationRepository
)

func readConfig(configsDirPath string) (config.Config, error) {
	config := config.Config{}

	configFilePath := filepath.Join(configsDirPath, "config.json")
	if !file.FileExists(configFilePath) {
		return config, fmt.Errorf("Config file does not exist: %s", configFilePath)
	}

	jsonContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(jsonContent, &config); err != nil {
		return config, err
	}

	return config, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slack_bot_hydration [flags]\n")
		flag.PrintDefaults()
	}

	pConfigsDirPath := flag.String("c", "", "Configs dir path")
	flag.Parse()

	if !file.FileExists(*pConfigsDirPath) {
		fmt.Println("Config dir path does not exist")
		return
	}

	configsDirPath := *pConfigsDirPath

	var err error
	appConfig, err := readConfig(configsDirPath)
	if err != nil {
		panic(err)
	}

	SetUp(appConfig)
	defer TearDown()

	e := echo.New()

	appLogFile, err := os.OpenFile(filepath.Join(appConfig.LogDirPath, "app.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetOutput(appLogFile)

	accessLogFile, err := os.OpenFile(filepath.Join(appConfig.LogDirPath, "access.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: accessLogFile,
	}))

	e.Use(middleware.Recover())

	e.POST("/", func(c echo.Context) error {
		return gateway(c, appConfig, configsDirPath)
	})

	e.Logger.Fatal(e.Start(appConfig.ServerHost))
}

// SetUp initializes App
func SetUp(appConfig config.Config) error {
	if appConfig.Db.Client == "postgresql" {
		repo = &repositories.HydrationPgRepository{}
	}

	err := repo.Connect(appConfig.Db)
	if err != nil {
		fmt.Println("Failed connecting db")
		return err
	}

	return nil
}

// TearDown destructs App
func TearDown() {
	repo.Close()
}

func gateway(c echo.Context, appConfig config.Config, configsDirPath string) error {
	payloadJSON := c.FormValue("payload")
	c.Echo().Logger.Info(payloadJSON)

	var payload interface{}
	err := json.Unmarshal([]byte(payloadJSON), &payload)
	if err != nil {
		c.Echo().Logger.Error(err)
		return c.String(http.StatusInternalServerError, "Error")
	}

	params, _ := json.Marshal(payload)
	c.Echo().Logger.Info(string(params))
	//prettyParams, _ := json.MarshalIndent(payload, "", "    ")
	//fmt.Println(string(prettyParams))

	iRequestType, err := jsonpointer.Get(payload, "/type")
	if err != nil {
		c.Echo().Logger.Error(err)
		return c.String(http.StatusInternalServerError, "Error")
	}
	requestType := iRequestType.(string)

	var iCallbackID interface{}
	switch requestType {
	case "shortcut":
		iCallbackID, _ = jsonpointer.Get(payload, "/callback_id")
	case "view_submission":
		iCallbackID, _ = jsonpointer.Get(payload, "/view/callback_id")
	case "block_actions":
		iCallbackID, _ = jsonpointer.Get(payload, "/actions/0/action_id")
	}

	callbackID := iCallbackID.(string)
	if len(callbackID) > 0 {
		switch callbackID {
		case "hydration__record_drink":
			return HandleOpenHydrationForm(c, appConfig, configsDirPath, payload)
		case "hydration__record_form":
			return HandleHydrationFormAddSubmission(c, appConfig, configsDirPath, payload)
		case "hydration__update_drink":
			return HandleOpenHydrationUpdateForm(c, appConfig, configsDirPath, payload)
		case "hydration__update_form":
			return HandleHydrationFormUpdateSubmission(c, appConfig, configsDirPath, payload)
		case "hydration__delete_drink":
			return HandleHydrationDelete(c, appConfig, configsDirPath, payload)
		case "hydration__repeat_drink":
			return HandleHydrationRepeat(c, appConfig, configsDirPath, payload)
		default:
			c.Echo().Logger.Warn("Unrecognized callbackID:", callbackID)
		}
	}

	return c.String(http.StatusForbidden, "Error")
}

// HandleOpenHydrationForm opens hydration record form modal.
func HandleOpenHydrationForm(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {

	// create goroutine for building modal and requesting view.open to Slack.
	go func() {
		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		triggerID, _ := jsonpointer.Get(payload, "/trigger_id")

		_, err := slackRepo.OpenHydrationAddView(triggerID.(string))
		if err != nil {
			c.Echo().Logger.Error(err)
		}
	}()

	return c.String(http.StatusOK, "Ok")
}

// HandleHydrationFormAddSubmission saves hydration and posts result message.
func HandleHydrationFormAddSubmission(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {

	iDrink, _ := jsonpointer.Get(payload, "/view/state/values/drink/drink/value")
	iAmount, _ := jsonpointer.Get(payload, "/view/state/values/amount/amount/selected_option/value")
	iUserName, _ := jsonpointer.Get(payload, "/user/username")

	drink, _ := iDrink.(string)
	amount, _ := iAmount.(string)
	userName, _ := iUserName.(string)

	intAmount, _ := strconv.ParseInt(amount, 10, 64)

	go func() {
		hydration := models.Hydration{
			Username: userName,
			Drink:    drink,
			Amount:   intAmount,
			Modified: time.Now(),
		}

		hydrationID, err := repo.Add(hydration)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}
		hydration.ID = hydrationID

		dailyAmount, err := repo.FetchDailyAmount(userName)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		_, err = slackRepo.PostHydrationAddResult(userName, "#general", hydration, dailyAmount)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}
	}()

	return c.String(http.StatusOK, "")
}

// HandleOpenHydrationUpdateForm opens hydration edit form modal.
func HandleOpenHydrationUpdateForm(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {
	iTriggerID, _ := jsonpointer.Get(payload, "/trigger_id")
	triggerID := iTriggerID.(string)

	iUserName, _ := jsonpointer.Get(payload, "/user/username")
	userName, _ := iUserName.(string)

	iHydrationID, _ := jsonpointer.Get(payload, "/actions/0/value")
	sHydrationID, _ := iHydrationID.(string)
	hydrationID, _ := strconv.ParseInt(sHydrationID, 10, 64)

	iMessageTS, _ := jsonpointer.Get(payload, "/container/message_ts")
	iChannel, _ := jsonpointer.Get(payload, "/container/channel_id")
	messageTS, _ := iMessageTS.(string)
	channel, _ := iChannel.(string)

	go func() {
		hydration, err := repo.FetchOne(hydrationID)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		if hydration.Username == userName {
			_, err = slackRepo.OpenHydrationUpdateView(triggerID, channel, messageTS, hydration)
			if err != nil {
				c.Echo().Logger.Error(err)
			}
		} else {
			_, err = slackRepo.ShowAlert(triggerID, "更新できません", "この記録を更新する権限がありません")
		}
	}()

	return c.String(http.StatusOK, "")
}

// HandleHydrationFormUpdateSubmission saves hydration and posts result message.
func HandleHydrationFormUpdateSubmission(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {

	iMetadata, _ := jsonpointer.Get(payload, "/view/private_metadata")
	iDrink, _ := jsonpointer.Get(payload, "/view/state/values/drink/drink/value")
	iAmount, _ := jsonpointer.Get(payload, "/view/state/values/amount/amount/selected_option/value")
	iUserName, _ := jsonpointer.Get(payload, "/user/username")

	metadata, _ := iMetadata.(string)
	metadataList := strings.Split(metadata, "-")

	channel := metadataList[0]
	messageTS := metadataList[1]
	hydrationID, _ := strconv.ParseInt(metadataList[2], 10, 64)

	drink, _ := iDrink.(string)
	amount, _ := iAmount.(string)
	userName, _ := iUserName.(string)

	intAmount, _ := strconv.ParseInt(amount, 10, 64)

	go func() {
		hydration := models.Hydration{
			ID:       hydrationID,
			Username: userName,
			Drink:    drink,
			Amount:   intAmount,
			Modified: time.Now(),
		}

		err := repo.Update(hydration)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		dailyAmount, err := repo.FetchDailyAmount(userName)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		_, err = slackRepo.PostHydrationUpdateResult(userName, channel, messageTS, hydration, dailyAmount)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}
	}()

	return c.String(http.StatusOK, "")
}

// HandleHydrationDelete deletes hydration and deletes message.
func HandleHydrationDelete(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {

	iHydrationID, _ := jsonpointer.Get(payload, "/actions/0/value")
	sHydrationID, _ := iHydrationID.(string)
	hydrationID, _ := strconv.ParseInt(sHydrationID, 10, 64)

	iUserName, _ := jsonpointer.Get(payload, "/user/username")
	userName, _ := iUserName.(string)

	iMessageTS, _ := jsonpointer.Get(payload, "/container/message_ts")
	messageTS, _ := iMessageTS.(string)

	iChannel, _ := jsonpointer.Get(payload, "/container/channel_id")
	channel, _ := iChannel.(string)

	go func() {
		hydration, err := repo.FetchOne(hydrationID)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		if hydration.Username == userName {
			err = repo.Delete(hydration)
			if err != nil {
				c.Echo().Logger.Error(err)
				return
			}

			_, err = slackRepo.DeleteMessage(channel, messageTS)
			if err != nil {
				c.Echo().Logger.Error(err)
			}
		} else {
			iTriggerID, _ := jsonpointer.Get(payload, "/trigger_id")
			triggerID := iTriggerID.(string)
			_, err = slackRepo.ShowAlert(triggerID, "削除できません", "この記録を削除する権限がありません")
		}

	}()

	return c.String(http.StatusOK, "")
}

// HandleHydrationRepeat adds same hydration from selected message.
func HandleHydrationRepeat(c echo.Context, appConfig config.Config, configsDirPath string, payload interface{}) error {

	iHydrationID, _ := jsonpointer.Get(payload, "/actions/0/value")
	sHydrationID, _ := iHydrationID.(string)
	hydrationID, _ := strconv.ParseInt(sHydrationID, 10, 64)

	iUserName, _ := jsonpointer.Get(payload, "/user/username")
	userName, _ := iUserName.(string)

	go func() {
		exHydration, err := repo.FetchOne(hydrationID)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		hydration := models.Hydration{
			Username: userName,
			Drink:    exHydration.Drink,
			Amount:   exHydration.Amount,
			Modified: time.Now(),
		}

		hydrationID, err := repo.Add(hydration)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}
		hydration.ID = hydrationID

		dailyAmount, err := repo.FetchDailyAmount(userName)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		_, err = slackRepo.PostHydrationAddResult(userName, "#general", hydration, dailyAmount)
		if err != nil {
			c.Echo().Logger.Error(err)
			return
		}

	}()

	return c.String(http.StatusOK, "")
}
