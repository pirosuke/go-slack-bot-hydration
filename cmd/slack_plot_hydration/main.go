package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pirosuke/slack-bot-hydration/internal/file"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/config"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/interfaces"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/repositories"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
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

	var repo interfaces.HydrationRepository
	if appConfig.Db.Client == "postgresql" {
		repo = &repositories.HydrationPgRepository{}
	}

	err = repo.Connect(appConfig.Db)
	if err != nil {
		fmt.Println("Failed connecting db")
		return
	}

	userList, err := repo.FetchWeeklyUsers()
	if err != nil {
		fmt.Println("Failed fetching user list")
		return
	}

	for _, userName := range userList {
		summaryList, err := repo.FetchWeeklySummary(userName)
		if err != nil {
			fmt.Println("Failed fetching summary for " + userName)
			fmt.Println(err)
			return
		}

		var dayList []string
		var amountList plotter.Values
		for _, summary := range summaryList {
			dayList = append(dayList, summary.Day)
			amountList = append(amountList, float64(summary.TotalAmount))
		}

		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Add(plotter.NewGrid())
		p.Title.Text = "週間水分摂取量"
		p.X.Label.Text = "曜日"
		p.Y.Label.Text = "飲んだ量"

		width := vg.Points(20)

		bars, err := plotter.NewBarChart(amountList, width)
		if err != nil {
			fmt.Println("Failed creating bar chart for " + userName)
			return
		}

		bars.LineStyle.Width = vg.Length(0)
		bars.Color = plotutil.Color(0)
		bars.Offset = -width

		p.Add(bars)
		p.NominalX(dayList...)

		outputFileName := "plot.png"
		outputPath := filepath.Join(appConfig.PlotOutputDirPath, outputFileName)
		if p.Save(5*vg.Inch, 3*vg.Inch, outputPath); err != nil {
			fmt.Println("Failed plot image output " + outputPath)
			return
		}

		slackRepo := &repositories.SlackRepository{
			Token:        appConfig.Slack.Token,
			ViewsDirPath: filepath.Join(configsDirPath, "views"),
		}

		_, err = slackRepo.UploadFile("#general", outputFileName, "png", outputPath, "過去一週間の水分摂取量を報告します")
		if err != nil {
			fmt.Println("Failed posting file to slack.")
			return
		}
	}
}
