package config

import (
	"github.com/pirosuke/slack-bot-hydration/internal/database"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/slack"
)

type (
	// Config describes global server config.
	Config struct {
		Db                database.DbConfig `json:"db"`
		ServerHost        string            `json:"host"`
		LogDirPath        string            `json:"log_dir"`
		PlotOutputDirPath string            `json:"plot_output_dir"`
		Slack             slack.Config      `json:"slack"`
	}
)
